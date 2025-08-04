# NAYSAYER Code Flow Diagram

A visual walkthrough of how NAYSAYER processes GitLab webhook requests and makes approval decisions.

## High-Level Flow

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   GitLab MR     │    │   NAYSAYER      │    │  YAML Analysis  │    │   Approval      │
│   Webhook       │───▶│   Webhook       │───▶│   Engine        │───▶│   Decision      │
│                 │    │   Handler       │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘
                              │                         │
                              ▼                         ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │   Payload       │    │   GitLab API    │
                       │   Validation    │    │   Calls         │
                       └─────────────────┘    └─────────────────┘
```

## Detailed Code Walkthrough

### 1. HTTP Request Processing

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                           INCOMING GITLAB WEBHOOK                                   │
└─────────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│  cmd/main.go:43                                                                    │
│  app.Post("/webhook", webhookHandler.HandleWebhook)                                │
└─────────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│  internal/handler/dataverse_product_config_review.go:32                           │
│  func (h *WebhookHandler) HandleWebhook(c *fiber.Ctx) error                       │
│                                                                                     │
│  Step 1: Parse JSON payload                                                        │
│  ├── c.BodyParser(&payload)                                                        │
│  └── Extract project ID and MR IID                                                 │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

### 2. Payload Validation & Extraction

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│  internal/gitlab/client.go:71                                                      │
│  func ExtractMRInfo(payload map[string]interface{}) (*MRInfo, error)              │
│                                                                                     │
│  Extract from GitLab webhook:                                                      │
│  ├── object_attributes.iid → MR IID (1551)                                         │
│  ├── project.id → Project ID (106670)                                              │
│  └── Validate both values exist                                                    │
└─────────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│  SUCCESS: Project=106670, MR=1551                                                  │
│  FAILURE: Return 400 Bad Request                                                   │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

### 3. Configuration Check

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│  internal/handler/dataverse_product_config_review.go:63                           │
│  func (h *WebhookHandler) analyzeFileChanges(projectID, mrIID int)                │
│                                                                                     │
│  Check GitLab Token:                                                               │
│  ├── if !h.config.HasGitLabToken() → Return "No Token" Decision                   │
│  └── Continue with analysis                                                        │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

### 4. GitLab API Integration

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│  internal/gitlab/client.go:28                                                      │
│  func (c *Client) FetchMRChanges(projectID, mrIID int) (*MRChanges, error)       │
│                                                                                     │
│  API Call:                                                                         │
│  GET /api/v4/projects/106670/merge_requests/1551/changes                          │
│  ├── Headers: Authorization: Bearer tDnsuUeVxy-n3PfhTvQG                          │
│  ├── Response: List of changed files with diffs                                   │
│  └── Returns: MRChanges struct                                                     │
└─────────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│  SUCCESS: Found changed files                                                      │
│  FAILURE: Return API Error Decision                                                │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

### 5. YAML Analysis (Single Path)

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│  internal/handler/dataverse_product_config_review.go:77                           │
│  Semantic YAML Analysis (Production Path)                                          │
└─────────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│  internal/analyzer/yaml.go:45                                                      │
│  func (a *YAMLAnalyzer) AnalyzeChanges(projectID, mrIID, changes)                 │
│                                                                                     │
│  For each changed file:                                                            │
│  ├── if !isProductYAML(file.path) → Skip                                          │
│  ├── if isProductYAML(file.path) → Analyze                                        │
│  └── Call analyzeFileChange(projectID, mrIID, filePath)                           │
└─────────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│  internal/analyzer/yaml.go:105                                                     │
│  func (a *YAMLAnalyzer) analyzeFileChange(projectID, mrIID, filePath)            │
│                                                                                     │
│  Step 1: Get target branch                                                         │
│  ├── GetMRTargetBranch(projectID, mrIID) → "main"                                 │
│                                                                                     │
│  Step 2: Fetch OLD file content                                                    │
│  ├── FetchFileContent(projectID, filePath, "main")                                │
│  ├── Base64 decode if needed                                                       │
│  └── Parse YAML → oldDP DataProduct                                                │
│                                                                                     │
│  Step 3: Get MR details                                                            │
│  ├── GetMRDetails(projectID, mrIID) → source_branch                               │
│                                                                                     │
│  Step 4: Fetch NEW file content                                                    │
│  ├── FetchFileContent(projectID, filePath, source_branch)                         │
│  ├── Base64 decode if needed                                                       │
│  └── Parse YAML → newDP DataProduct                                                │
│                                                                                     │
│  Step 5: Compare warehouses                                                        │
│  └── compareWarehouses(filePath, oldDP, newDP)                                     │
└─────────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                               SUCCESS / FAILURE                                    │
│  ┌─────────────────────┐              │              ┌─────────────────────┐      │
│  │      SUCCESS        │              │              │      FAILURE        │      │
│  │ Return warehouse    │              │              │ Return explicit     │      │
│  │ changes detected    │              │              │ AnalysisError       │      │
│  └─────────────────────┘              │              └─────────────────────┘      │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

### 6. Warehouse Comparison Logic

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│  internal/analyzer/yaml.go:161                                                     │
│  func (a *YAMLAnalyzer) compareWarehouses(filePath, oldDP, newDP)                 │
│                                                                                     │
│  Create comparison maps:                                                            │
│  ├── oldWarehouses["user"] = "XSMALL"                                             │
│  ├── oldWarehouses["service_account"] = "XSMALL"                                  │
│  ├── newWarehouses["user"] = "XSMALL"                                             │
│  └── newWarehouses["service_account"] = "LARGE"                                   │
│                                                                                     │
│  Detect changes:                                                                    │
│  ├── For each warehouse type in newWarehouses:                                     │
│  ├── if oldSize != newSize:                                                        │
│  │   ├── oldValue = WarehouseSizes["XSMALL"] = 1                                  │
│  │   ├── newValue = WarehouseSizes["LARGE"] = 4                                   │
│  │   ├── isDecrease = (oldValue > newValue) = (1 > 4) = false                    │
│  │   └── Create WarehouseChange{FromSize: "XSMALL", ToSize: "LARGE", ...}         │
│  └── Return []WarehouseChange                                                      │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

### 7. Decision Making

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│  internal/decision/maker.go:14                                                     │
│  func (m *Maker) Decide(changes []WarehouseChange) Decision                       │
│                                                                                     │
│  Decision Logic:                                                                    │
│  ├── if len(changes) == 0 → "no warehouse changes detected"                       │
│  ├── for each change:                                                              │
│  │   └── if !change.IsDecrease → "warehouse increase detected"                    │
│  └── if all decreases → "all warehouse changes are decreases" ✅                  │
└─────────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                              EXAMPLE DECISIONS                                     │
│                                                                                     │
│  Scenario 1: XSMALL → LARGE (increase)                                            │
│  ┌─────────────────────────────────────────────────────────────────────────────┐  │
│  │ Decision{                                                                   │  │
│  │   AutoApprove: false,                                                       │  │
│  │   Reason: "warehouse increase detected: XSMALL → LARGE",                   │  │
│  │   Summary: "🚫 Warehouse increase - platform approval required",           │  │
│  │   Details: "File: dataproducts/.../product.yaml (type: service_account)"  │  │
│  │ }                                                                           │  │
│  └─────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                     │
│  Scenario 2: LARGE → MEDIUM (decrease)                                            │
│  ┌─────────────────────────────────────────────────────────────────────────────┐  │
│  │ Decision{                                                                   │  │
│  │   AutoApprove: true,                                                        │  │
│  │   Reason: "all warehouse changes are decreases",                           │  │
│  │   Summary: "✅ Warehouse decrease(s) - auto-approved",                     │  │
│  │   Details: "Found 1 warehouse decrease(s)"                                 │  │
│  │ }                                                                           │  │
│  └─────────────────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

### 8. HTTP Response

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│  internal/handler/dataverse_product_config_review.go:85                           │
│  return c.JSON(decision)                                                           │
│                                                                                     │
│  HTTP Response:                                                                     │
│  ├── Status: 200 OK                                                                │
│  ├── Content-Type: application/json                                                │
│  └── Body: Decision JSON                                                           │
└─────────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                              GITLAB RECEIVES RESPONSE                              │
│                                                                                     │
│  {                                                                                  │
│    "auto_approve": false,                                                          │
│    "reason": "warehouse increase detected: XSMALL → LARGE",                       │
│    "summary": "🚫 Warehouse increase - platform approval required",               │
│    "details": "File: dataproducts/aggregate/discounting/preprod/product.yaml..."  │
│  }                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

## Error Flow Scenarios

### Scenario A: Missing GitLab Token

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Webhook       │    │   Config Check  │    │   Error         │
│   Received      │───▶│   No Token      │───▶│   Response      │
│                 │    │   Detected      │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │ Return Decision │
                       │ auto_approve:   │
                       │   false         │
                       │ reason: "GitLab │
                       │ token not       │
                       │ configured"     │
                       └─────────────────┘
```

### Scenario B: GitLab API Failure

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Webhook       │    │   API Call      │    │   API Error     │    │   Error         │
│   Received      │───▶│   to GitLab     │───▶│   (401, 404,    │───▶│   Response      │
│                 │    │                 │    │    500, etc.)   │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘
                                                        │
                                                        ▼
                                               ┌─────────────────┐
                                               │ Return Decision │
                                               │ auto_approve:   │
                                               │   false         │
                                               │ reason: "Failed │
                                               │ to fetch file   │
                                               │ changes"        │
                                               └─────────────────┘
```

### Scenario C: YAML Analysis Fails → Explicit Error

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   YAML Analysis │    │   File Not      │    │   Log Error     │    │   Return        │
│   Attempt       │───▶│   Found or API  │───▶│   Return        │───▶│   AnalysisError │
│                 │    │   Failure       │    │   Error         │    │   Decision      │
└─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘
                                                        │
                                                        ▼
                                               ┌─────────────────┐
                                               │ Requires manual │
                                               │ approval with   │
                                               │ error details   │
                                               └─────────────────┘
```

## File Structure Impact

```
File Processing Priority:
┌─────────────────────────────────────────┐
│  product.yaml files                     │ ✅ ANALYZED
│  ├── dataproducts/source/*/env/         │
│  ├── dataproducts/aggregate/*/env/      │
│  └── dataproducts/platform/*/env/       │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│  Other files                            │ ❌ IGNORED
│  ├── README.md                          │
│  ├── scripts/*.sh                       │
│  ├── *.json                             │
│  └── Other *.yaml (not product.yaml)    │
└─────────────────────────────────────────┘
```

## Performance Characteristics

```
Request Flow Timeline:
┌─────────┬─────────┬─────────┬─────────┬─────────┬─────────┐
│ 0ms    │ 50ms    │ 200ms   │ 500ms   │ 1000ms  │ 2000ms  │
├─────────┼─────────┼─────────┼─────────┼─────────┼─────────┤
│ Webhook │ Parse   │ GitLab  │ YAML    │ Decision│ Response│
│ Received│ Payload │ API     │ Analysis│ Logic   │ Sent    │
│         │         │ Calls   │         │         │         │
└─────────┴─────────┴─────────┴─────────┴─────────┴─────────┘

GitLab API Calls per Request:
├── GET /merge_requests/{iid}/changes     (1 call)
├── GET /merge_requests/{iid}             (1 call)  
├── GET /repository/files/{path}?ref=main (1 call per file)
└── GET /repository/files/{path}?ref=feat (1 call per file)

Total: 2 + (2 × number_of_product_yaml_files)
```

This flow diagram shows exactly how NAYSAYER processes each webhook request, from the initial HTTP request through the final JSON response, including all error scenarios and fallback mechanisms.