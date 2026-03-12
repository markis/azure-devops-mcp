# Fields by Work Item Type - Summary

**Source:** fields-by-type.json  
**Generated:** 2026-03-11  
**Project:** Access Analyzer

## Key Findings

### Common Fields (on ALL 24 types)

**48 fields** are available on every single work item type, including:

**Essential System Fields:**

- `System.Title` [REQUIRED on all types]
- `System.State` [REQUIRED on all types]
- `System.AreaId` [REQUIRED on all types]
- `System.IterationId` [REQUIRED on all types]
- `System.Description`
- `System.AssignedTo`
- `System.Tags`
- `System.AreaPath`
- `System.IterationPath`
- `System.Reason`
- `System.History`
- `System.Parent`

**Metadata Fields (read-only but available):**

- `System.CreatedBy`, `System.CreatedDate`
- `System.ChangedBy`, `System.ChangedDate`
- `System.Rev`, `System.RevisedDate`
- `System.BoardColumn`, `System.BoardColumnDone`, `System.BoardLane`
- Various count fields (AttachedFileCount, CommentCount, ExternalLinkCount, etc.)

### Type-Specific Field Summary

| Work Item Type | Total Fields | Required | Notable Type-Specific Fields |
|----------------|--------------|----------|------------------------------|
| **User Story** | 74 | 5 | AcceptanceCriteria, StoryPoints, Risk, BusinessCase, DevOwner, Poker |
| **Task** | 66 | 4 | Activity, OriginalEstimate, RemainingWork, CompletedWork |
| **Bug** | 78 | 5 | Severity, ReproSteps (via TCM), FoundIn, RCAReason, SalesforceCaseNumber |
| **Feature** | 86 | 8 | BusinessValue, Risk, Effort, TimeCriticality, MarketDate, QCStartDate, Documentation [REQ], TeeshirtSizing [REQ], AtRisk [REQ] |
| **Epic** | 66 | 6 | BusinessValue, Risk, Effort, TimeCriticality, Priority [REQ] |
| **Defect** | 68 | 4 | Severity, ClosedInBuild, TimeSpent, RCAReason |
| **Vulnerability** | 68 | 4 | Severity, CVENumber, VulnerabilitySource, TimeSpent |
| **Security** | 70 | 4 | CVENumber, VulnerabilitySource, TeeshirtSizing, Confidence, DeliveryRisk |
| **Escalation** | 75 | 7 | SalesforceCaseNumber, SalesforceURL, SalesforceEscalatingEngineer, DaysAwaitingDev, DaysAwaitingSupport, InitialDetailQuality, PrioritizationScore |
| **Technical Task** | 68 | 5 | Effort, TeeshirtSizing, AtRisk, Confidence, DeliveryRisk, MitigationPlan |
| **Issue** | 63 | 4 | DueDate |
| **Test Case** | 66 | 4 | Steps, AutomatedTestName, AutomatedTestStorage, AutomationStatus, Parameters |

## Fields by Category

### ✅ Already Implemented in MCP

#### System Fields

- ✅ `System.Title`
- ✅ `System.State`
- ✅ `System.Description`
- ✅ `System.AssignedTo`
- ✅ `System.Tags`
- ✅ `System.AreaPath`
- ✅ `System.IterationPath`
- ✅ `System.Reason`

#### Microsoft.VSTS.Common.*

- ✅ `Priority` - On most types
- ✅ `Severity` - On Bug, Defect, Vulnerability
- ✅ `AcceptanceCriteria` - On Feature, User Story, Escalation
- ✅ `Activity` - On Task, Bug, Defect types
- ✅ `ValueArea` - On Epic, Feature, User Story, Bug, Defect, Vulnerability, Security, Escalation

#### Microsoft.VSTS.Scheduling.*

- ✅ `StoryPoints` - On User Story, Feature, Bug
- ✅ `OriginalEstimate` - On Task, Bug, User Story
- ✅ `RemainingWork` - On Task, Bug, Feature
- ✅ `CompletedWork` - On Task, Bug, User Story
- ✅ `Effort` - On Epic, Feature

#### Custom Fields

- ✅ `Custom.Teeshirtsizing` - On Feature, Security, Technical Task

### ❌ NOT Implemented but Available

#### Date/Time Fields (Microsoft.VSTS.Scheduling.*)

- ❌ `StartDate` - On Task, Epic, Feature, User Story, Test Plan
- ❌ `FinishDate` - On Task, User Story, Test Plan
- ❌ `TargetDate` - On Epic, Feature
- ❌ `DueDate` - On Issue only

#### Status Tracking (Microsoft.VSTS.Common.*)

- ❌ `ActivatedBy` - On most types
- ❌ `ActivatedDate` - On most types
- ❌ `ResolvedBy` - On Bug, User Story, Epic, Feature, Issue
- ❌ `ResolvedDate` - On Bug, User Story, Epic, Feature, Issue
- ❌ `ResolvedReason` - On Bug, User Story, Epic, Feature
- ❌ `ClosedBy` - On most types
- ❌ `ClosedDate` - On most types
- ❌ `StateChangeDate` - On most types

#### Planning/Prioritization (Microsoft.VSTS.Common.*)

- ❌ `BusinessValue` - On Epic, Feature (important for prioritization!)
- ❌ `Risk` - On Epic, Feature, User Story
- ❌ `StackRank` - On all main types (backlog ordering!)
- ❌ `TimeCriticality` - On Epic, Feature

#### Build Integration (Microsoft.VSTS.Build.*)

- ❌ `FoundIn` - On Bug only
- ❌ `IntegrationBuild` - None found in current types

#### Test Management (Microsoft.VSTS.TCM.*)

- ❌ `ReproSteps` - On Bug (currently supported!)
- ❌ `Steps` - On Test Case, Shared Steps
- ❌ `AutomatedTestName` - On Test Case
- ❌ `AutomatedTestStorage` - On Test Case
- ❌ `AutomatedTestType` - On Test Case
- ❌ `AutomatedTestId` - On Test Case
- ❌ `AutomationStatus` - On Test Case
- ❌ `Parameters` - On Test Case, Shared Steps, Shared Parameter
- ❌ `SystemInfo` - Not found on Bug (surprising!)

#### CMMI Fields (Microsoft.VSTS.CMMI.*)

- ❌ `Blocked` - Not found (may not be in Agile template)
- ❌ `Comments` - On Test Case only
- ❌ `MitigationPlan` - On Feature, Technical Task

#### Code Review (Microsoft.VSTS.CodeReview.*)

- ❌ `Context` - On Code Review Request
- ❌ `AcceptedBy` - On Code Review Response
- ❌ `AcceptedDate` - On Code Review Response
- ❌ `ClosedStatus` - On Code Review Request/Response
- ❌ `ClosingComment` - On Code Review Request/Response

## Required Fields by Type

Fields marked as "alwaysRequired" for each work item type:

### User Story (5 required)

1. `System.AreaId`
2. `System.IterationId`
3. `System.State`
4. `System.Title`
5. `Microsoft.VSTS.Common.ValueArea`

### Task (4 required)

1. `System.AreaId`
2. `System.IterationId`
3. `System.State`
4. `System.Title`

### Bug (5 required)

1. `System.AreaId`
2. `System.IterationId`
3. `System.State`
4. `System.Title`
5. `Microsoft.VSTS.Common.ValueArea`

### Feature (8 required) ⚠️ Most required fields

1. `System.AreaId`
2. `System.IterationId`
3. `System.State`
4. `System.Title`
5. `Microsoft.VSTS.Common.ValueArea`
6. `Custom.AtRisk`
7. `Custom.Documentation`
8. `Custom.Teeshirtsizing`

### Epic (6 required)

1. `System.AreaId`
2. `System.IterationId`
3. `System.State`
4. `System.Title`
5. `Microsoft.VSTS.Common.Priority`
6. `Microsoft.VSTS.Common.ValueArea`

### Escalation (7 required)

1. `System.AreaId`
2. `System.IterationId`
3. `System.State`
4. `System.Title`
5. `Microsoft.VSTS.Common.ValueArea`
6. Plus 2 custom required fields

## Recommendations

### 🔴 High Priority - Add These Fields

These are commonly used and provide significant value:

1. **Date Fields** (essential for planning)
   - `Microsoft.VSTS.Scheduling.StartDate`
   - `Microsoft.VSTS.Scheduling.FinishDate`
   - `Microsoft.VSTS.Scheduling.TargetDate`
   - `Microsoft.VSTS.Scheduling.DueDate`

2. **Status Tracking** (essential for workflow visibility)
   - `Microsoft.VSTS.Common.ActivatedDate`
   - `Microsoft.VSTS.Common.ResolvedDate`
   - `Microsoft.VSTS.Common.ClosedDate`
   - `Microsoft.VSTS.Common.StateChangeDate`

3. **Planning/Prioritization** (important for backlog management)
   - `Microsoft.VSTS.Common.BusinessValue` ⭐ Key for Features/Epics
   - `Microsoft.VSTS.Common.StackRank` ⭐ Critical for backlog ordering
   - `Microsoft.VSTS.Common.Risk`
   - `Microsoft.VSTS.Common.TimeCriticality`

4. **Build Integration** (useful for release tracking)
   - `Microsoft.VSTS.Build.FoundIn`

5. **Custom Required Fields** (needed for Feature creation)
   - `Custom.AtRisk` (boolean) - REQUIRED on Feature
   - `Custom.Documentation` (picklist) - REQUIRED on Feature
   - Note: `Custom.Teeshirtsizing` already implemented

### 🟡 Medium Priority - Consider Adding

1. **Identity Tracking**
   - `Microsoft.VSTS.Common.ActivatedBy`
   - `Microsoft.VSTS.Common.ResolvedBy`
   - `Microsoft.VSTS.Common.ClosedBy`

2. **Custom Date Fields** (Netwrix-specific planning)
   - `Custom.MarketDate`
   - `Custom.DevCompleteDate`
   - `Custom.QCStartDate`
   - `Custom.QCCompleteDate`

3. **Salesforce Integration** (if used in workflows)
   - `Custom.SalesforceCaseNumber`
   - `Custom.SalesforceURL`
   - `Custom.SalesforceEscalatingEngineer`

### 🟢 Low Priority - Specialized Use Cases

1. **Test Management Fields** (only if supporting test automation)
   - `Microsoft.VSTS.TCM.Steps`
   - `Microsoft.VSTS.TCM.AutomatedTestName`
   - `Microsoft.VSTS.TCM.AutomationStatus`

2. **Code Review Fields** (only if supporting code review workflows)
    - `Microsoft.VSTS.CodeReview.*` fields

3. **Advanced Custom Fields** (domain-specific)
    - `Custom.CVENumber` (security)
    - `Custom.RCAReason` (quality)
    - `Custom.Poker` (planning poker)

## Implementation Notes

### Field Types to Support

Based on the discovery, you'll need to handle these types:

- ✅ `string` - Already supported
- ✅ `integer` - Already supported (Priority)
- ✅ `double` - Already supported (StoryPoints, etc.)
- ✅ `html` - Already supported (Description, AcceptanceCriteria)
- ✅ `plainText` - Similar to string, already supported (Tags)
- ✅ `treePath` - Already supported (AreaPath, IterationPath)
- ❌ `dateTime` - **Need to add support**
- ❌ `boolean` - **Need to add support**
- ✅ `history` - Already supported (History field)

### Picklist Fields

The following custom fields have allowed values (picklists):

- `Custom.DeliveryRisk`
- `Custom.Documentation`
- `Custom.EscalationCloseReason`
- `Custom.InitialDetailQuality`
- `Custom.Poker`
- `Custom.RCAReason`
- `Custom.Teeshirtsizing` (already implemented)
- `Custom.VulnerabilitySource`

**Recommendation:** Expose allowed values in schema descriptions to help Claude choose valid options.

### System Fields (Read-Only)

These are available but not editable - consider exposing for context:

- `System.CreatedBy`, `System.CreatedDate`
- `System.ChangedBy`, `System.ChangedDate`
- `System.Id`, `System.Rev`
- `System.BoardColumn`, `System.BoardLane`

## Next Steps

1. **Add DateTime support** - Required for date fields
2. **Add Boolean support** - Required for `Custom.AtRisk`, etc.
3. **Implement high-priority fields** from recommendations
4. **Consider exposing read-only fields** for context
5. **Add picklist validation** - Include allowed values in schemas
6. **Update error messages** - Include new fields in `ErrNoFieldsToUpdate`
