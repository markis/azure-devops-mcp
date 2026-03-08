//nolint:lll,revive // Generated mock with long lines and many methods
package test

import (
	"context"
	"io"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"
)

// mockWITClient implements workitemtracking.Client for testing.
// Only implements methods used by RealADOClient.
type mockWITClient struct {
	GetWorkItemFn       func(context.Context, workitemtracking.GetWorkItemArgs) (*workitemtracking.WorkItem, error)
	GetWorkItemsBatchFn func(context.Context, workitemtracking.GetWorkItemsBatchArgs) (*[]workitemtracking.WorkItem, error)
	QueryByWiqlFn       func(context.Context, workitemtracking.QueryByWiqlArgs) (*workitemtracking.WorkItemQueryResult, error)
	CreateWorkItemFn    func(context.Context, workitemtracking.CreateWorkItemArgs) (*workitemtracking.WorkItem, error)
	UpdateWorkItemFn    func(context.Context, workitemtracking.UpdateWorkItemArgs) (*workitemtracking.WorkItem, error)
	AddCommentFn        func(context.Context, workitemtracking.AddCommentArgs) (*workitemtracking.Comment, error)
}

// GetWorkItem delegates to GetWorkItemFn.
func (m *mockWITClient) GetWorkItem(ctx context.Context, args workitemtracking.GetWorkItemArgs) (*workitemtracking.WorkItem, error) {
	if m.GetWorkItemFn != nil {
		return m.GetWorkItemFn(ctx, args)
	}

	panic("GetWorkItemFn not set")
}

// QueryByWiql delegates to QueryByWiqlFn.
func (m *mockWITClient) QueryByWiql(ctx context.Context, args workitemtracking.QueryByWiqlArgs) (*workitemtracking.WorkItemQueryResult, error) {
	if m.QueryByWiqlFn != nil {
		return m.QueryByWiqlFn(ctx, args)
	}

	panic("QueryByWiqlFn not set")
}

// CreateWorkItem delegates to CreateWorkItemFn.
func (m *mockWITClient) CreateWorkItem(ctx context.Context, args workitemtracking.CreateWorkItemArgs) (*workitemtracking.WorkItem, error) {
	if m.CreateWorkItemFn != nil {
		return m.CreateWorkItemFn(ctx, args)
	}

	panic("CreateWorkItemFn not set")
}

// UpdateWorkItem delegates to UpdateWorkItemFn.
func (m *mockWITClient) UpdateWorkItem(ctx context.Context, args workitemtracking.UpdateWorkItemArgs) (*workitemtracking.WorkItem, error) {
	if m.UpdateWorkItemFn != nil {
		return m.UpdateWorkItemFn(ctx, args)
	}

	panic("UpdateWorkItemFn not set")
}

// AddComment delegates to AddCommentFn.
func (m *mockWITClient) AddComment(ctx context.Context, args workitemtracking.AddCommentArgs) (*workitemtracking.Comment, error) {
	if m.AddCommentFn != nil {
		return m.AddCommentFn(ctx, args)
	}

	panic("AddCommentFn not set")
}

// All other workitemtracking.Client methods panic with "not implemented"

func (m *mockWITClient) AddWorkItemComment(context.Context, workitemtracking.AddWorkItemCommentArgs) (*workitemtracking.Comment, error) {
	panic("AddWorkItemComment not implemented in mock")
}

func (m *mockWITClient) CreateAttachment(context.Context, workitemtracking.CreateAttachmentArgs) (*workitemtracking.AttachmentReference, error) {
	panic("CreateAttachment not implemented in mock")
}

func (m *mockWITClient) CreateCommentReaction(context.Context, workitemtracking.CreateCommentReactionArgs) (*workitemtracking.CommentReaction, error) {
	panic("CreateCommentReaction not implemented in mock")
}

func (m *mockWITClient) CreateOrUpdateClassificationNode(context.Context, workitemtracking.CreateOrUpdateClassificationNodeArgs) (*workitemtracking.WorkItemClassificationNode, error) {
	panic("CreateOrUpdateClassificationNode not implemented in mock")
}

func (m *mockWITClient) CreateQuery(context.Context, workitemtracking.CreateQueryArgs) (*workitemtracking.QueryHierarchyItem, error) {
	panic("CreateQuery not implemented in mock")
}

func (m *mockWITClient) CreateTemplate(context.Context, workitemtracking.CreateTemplateArgs) (*workitemtracking.WorkItemTemplate, error) {
	panic("CreateTemplate not implemented in mock")
}

func (m *mockWITClient) CreateTempQuery(context.Context, workitemtracking.CreateTempQueryArgs) (*workitemtracking.TemporaryQueryResponseModel, error) {
	panic("CreateTempQuery not implemented in mock")
}

func (m *mockWITClient) CreateWorkItemField(context.Context, workitemtracking.CreateWorkItemFieldArgs) (*workitemtracking.WorkItemField2, error) {
	panic("CreateWorkItemField not implemented in mock")
}

func (m *mockWITClient) DeleteClassificationNode(context.Context, workitemtracking.DeleteClassificationNodeArgs) error {
	panic("DeleteClassificationNode not implemented in mock")
}

func (m *mockWITClient) DeleteComment(context.Context, workitemtracking.DeleteCommentArgs) error {
	panic("DeleteComment not implemented in mock")
}

func (m *mockWITClient) DeleteCommentReaction(context.Context, workitemtracking.DeleteCommentReactionArgs) (*workitemtracking.CommentReaction, error) {
	panic("DeleteCommentReaction not implemented in mock")
}

func (m *mockWITClient) DeleteQuery(context.Context, workitemtracking.DeleteQueryArgs) error {
	panic("DeleteQuery not implemented in mock")
}

func (m *mockWITClient) DeleteTag(context.Context, workitemtracking.DeleteTagArgs) error {
	panic("DeleteTag not implemented in mock")
}

func (m *mockWITClient) DeleteTemplate(context.Context, workitemtracking.DeleteTemplateArgs) error {
	panic("DeleteTemplate not implemented in mock")
}

func (m *mockWITClient) DeleteWorkItem(context.Context, workitemtracking.DeleteWorkItemArgs) (*workitemtracking.WorkItemDelete, error) {
	panic("DeleteWorkItem not implemented in mock")
}

func (m *mockWITClient) DeleteWorkItemField(context.Context, workitemtracking.DeleteWorkItemFieldArgs) error {
	panic("DeleteWorkItemField not implemented in mock")
}

func (m *mockWITClient) DeleteWorkItems(context.Context, workitemtracking.DeleteWorkItemsArgs) (*workitemtracking.WorkItemDeleteBatch, error) {
	panic("DeleteWorkItems not implemented in mock")
}

func (m *mockWITClient) DestroyWorkItem(context.Context, workitemtracking.DestroyWorkItemArgs) error {
	panic("DestroyWorkItem not implemented in mock")
}

func (m *mockWITClient) GetAttachmentContent(context.Context, workitemtracking.GetAttachmentContentArgs) (io.ReadCloser, error) {
	panic("GetAttachmentContent not implemented in mock")
}

func (m *mockWITClient) GetAttachmentZip(context.Context, workitemtracking.GetAttachmentZipArgs) (io.ReadCloser, error) {
	panic("GetAttachmentZip not implemented in mock")
}

func (m *mockWITClient) GetClassificationNode(context.Context, workitemtracking.GetClassificationNodeArgs) (*workitemtracking.WorkItemClassificationNode, error) {
	panic("GetClassificationNode not implemented in mock")
}

func (m *mockWITClient) GetClassificationNodes(context.Context, workitemtracking.GetClassificationNodesArgs) (*[]workitemtracking.WorkItemClassificationNode, error) {
	panic("GetClassificationNodes not implemented in mock")
}

func (m *mockWITClient) GetComment(context.Context, workitemtracking.GetCommentArgs) (*workitemtracking.Comment, error) {
	panic("GetComment not implemented in mock")
}

func (m *mockWITClient) GetCommentReactions(context.Context, workitemtracking.GetCommentReactionsArgs) (*[]workitemtracking.CommentReaction, error) {
	panic("GetCommentReactions not implemented in mock")
}

func (m *mockWITClient) GetComments(context.Context, workitemtracking.GetCommentsArgs) (*workitemtracking.CommentList, error) {
	panic("GetComments not implemented in mock")
}

func (m *mockWITClient) GetCommentsBatch(context.Context, workitemtracking.GetCommentsBatchArgs) (*workitemtracking.CommentList, error) {
	panic("GetCommentsBatch not implemented in mock")
}

func (m *mockWITClient) GetCommentVersion(context.Context, workitemtracking.GetCommentVersionArgs) (*workitemtracking.CommentVersion, error) {
	panic("GetCommentVersion not implemented in mock")
}

func (m *mockWITClient) GetCommentVersions(context.Context, workitemtracking.GetCommentVersionsArgs) (*[]workitemtracking.CommentVersion, error) {
	panic("GetCommentVersions not implemented in mock")
}

func (m *mockWITClient) GetDeletedWorkItem(context.Context, workitemtracking.GetDeletedWorkItemArgs) (*workitemtracking.WorkItemDelete, error) {
	panic("GetDeletedWorkItem not implemented in mock")
}

func (m *mockWITClient) GetDeletedWorkItems(context.Context, workitemtracking.GetDeletedWorkItemsArgs) (*[]workitemtracking.WorkItemDeleteReference, error) {
	panic("GetDeletedWorkItems not implemented in mock")
}

func (m *mockWITClient) GetDeletedWorkItemShallowReferences(context.Context, workitemtracking.GetDeletedWorkItemShallowReferencesArgs) (*[]workitemtracking.WorkItemDeleteShallowReference, error) {
	panic("GetDeletedWorkItemShallowReferences not implemented in mock")
}

func (m *mockWITClient) GetEngagedUsers(context.Context, workitemtracking.GetEngagedUsersArgs) (*[]webapi.IdentityRef, error) {
	panic("GetEngagedUsers not implemented in mock")
}

func (m *mockWITClient) GetGithubConnectionRepositories(context.Context, workitemtracking.GetGithubConnectionRepositoriesArgs) (*[]workitemtracking.GitHubConnectionRepoModel, error) {
	panic("GetGithubConnectionRepositories not implemented in mock")
}

func (m *mockWITClient) GetGithubConnections(context.Context, workitemtracking.GetGithubConnectionsArgs) (*[]workitemtracking.GitHubConnectionModel, error) {
	panic("GetGithubConnections not implemented in mock")
}

func (m *mockWITClient) GetQueries(context.Context, workitemtracking.GetQueriesArgs) (*[]workitemtracking.QueryHierarchyItem, error) {
	panic("GetQueries not implemented in mock")
}

func (m *mockWITClient) GetQueriesBatch(context.Context, workitemtracking.GetQueriesBatchArgs) (*[]workitemtracking.QueryHierarchyItem, error) {
	panic("GetQueriesBatch not implemented in mock")
}

func (m *mockWITClient) GetQuery(context.Context, workitemtracking.GetQueryArgs) (*workitemtracking.QueryHierarchyItem, error) {
	panic("GetQuery not implemented in mock")
}

func (m *mockWITClient) GetQueryResultCount(context.Context, workitemtracking.GetQueryResultCountArgs) (*int, error) {
	panic("GetQueryResultCount not implemented in mock")
}

func (m *mockWITClient) GetRecentActivityData(context.Context, workitemtracking.GetRecentActivityDataArgs) (*[]workitemtracking.AccountRecentActivityWorkItemModel2, error) {
	panic("GetRecentActivityData not implemented in mock")
}

func (m *mockWITClient) GetRelationType(context.Context, workitemtracking.GetRelationTypeArgs) (*workitemtracking.WorkItemRelationType, error) {
	panic("GetRelationType not implemented in mock")
}

func (m *mockWITClient) GetRelationTypes(context.Context, workitemtracking.GetRelationTypesArgs) (*[]workitemtracking.WorkItemRelationType, error) {
	panic("GetRelationTypes not implemented in mock")
}

func (m *mockWITClient) GetReportingLinksByLinkType(context.Context, workitemtracking.GetReportingLinksByLinkTypeArgs) (*workitemtracking.ReportingWorkItemLinksBatch, error) {
	panic("GetReportingLinksByLinkType not implemented in mock")
}

func (m *mockWITClient) GetRevision(context.Context, workitemtracking.GetRevisionArgs) (*workitemtracking.WorkItem, error) {
	panic("GetRevision not implemented in mock")
}

func (m *mockWITClient) GetRevisions(context.Context, workitemtracking.GetRevisionsArgs) (*[]workitemtracking.WorkItem, error) {
	panic("GetRevisions not implemented in mock")
}

func (m *mockWITClient) GetRootNodes(context.Context, workitemtracking.GetRootNodesArgs) (*[]workitemtracking.WorkItemClassificationNode, error) {
	panic("GetRootNodes not implemented in mock")
}

func (m *mockWITClient) GetTag(context.Context, workitemtracking.GetTagArgs) (*workitemtracking.WorkItemTagDefinition, error) {
	panic("GetTag not implemented in mock")
}

func (m *mockWITClient) GetTags(context.Context, workitemtracking.GetTagsArgs) (*[]workitemtracking.WorkItemTagDefinition, error) {
	panic("GetTags not implemented in mock")
}

func (m *mockWITClient) GetTemplate(context.Context, workitemtracking.GetTemplateArgs) (*workitemtracking.WorkItemTemplate, error) {
	panic("GetTemplate not implemented in mock")
}

func (m *mockWITClient) GetTemplates(context.Context, workitemtracking.GetTemplatesArgs) (*[]workitemtracking.WorkItemTemplateReference, error) {
	panic("GetTemplates not implemented in mock")
}

func (m *mockWITClient) GetUpdate(context.Context, workitemtracking.GetUpdateArgs) (*workitemtracking.WorkItemUpdate, error) {
	panic("GetUpdate not implemented in mock")
}

func (m *mockWITClient) GetUpdates(context.Context, workitemtracking.GetUpdatesArgs) (*[]workitemtracking.WorkItemUpdate, error) {
	panic("GetUpdates not implemented in mock")
}

func (m *mockWITClient) GetWorkArtifactLinkTypes(context.Context, workitemtracking.GetWorkArtifactLinkTypesArgs) (*[]workitemtracking.WorkArtifactLink, error) {
	panic("GetWorkArtifactLinkTypes not implemented in mock")
}

func (m *mockWITClient) GetWorkItemField(context.Context, workitemtracking.GetWorkItemFieldArgs) (*workitemtracking.WorkItemField2, error) {
	panic("GetWorkItemField not implemented in mock")
}

func (m *mockWITClient) GetWorkItemFields(context.Context, workitemtracking.GetWorkItemFieldsArgs) (*[]workitemtracking.WorkItemField2, error) {
	panic("GetWorkItemFields not implemented in mock")
}

func (m *mockWITClient) GetWorkItemIconJson(context.Context, workitemtracking.GetWorkItemIconJsonArgs) (*workitemtracking.WorkItemIcon, error) {
	panic("GetWorkItemIconJson not implemented in mock")
}

func (m *mockWITClient) GetWorkItemIcons(context.Context, workitemtracking.GetWorkItemIconsArgs) (*[]workitemtracking.WorkItemIcon, error) {
	panic("GetWorkItemIcons not implemented in mock")
}

func (m *mockWITClient) GetWorkItemIconSvg(context.Context, workitemtracking.GetWorkItemIconSvgArgs) (io.ReadCloser, error) {
	panic("GetWorkItemIconSvg not implemented in mock")
}

func (m *mockWITClient) GetWorkItemIconXaml(context.Context, workitemtracking.GetWorkItemIconXamlArgs) (io.ReadCloser, error) {
	panic("GetWorkItemIconXaml not implemented in mock")
}

func (m *mockWITClient) GetWorkItemNextStatesOnCheckinAction(context.Context, workitemtracking.GetWorkItemNextStatesOnCheckinActionArgs) (*[]workitemtracking.WorkItemNextStateOnTransition, error) {
	panic("GetWorkItemNextStatesOnCheckinAction not implemented in mock")
}

func (m *mockWITClient) GetWorkItems(context.Context, workitemtracking.GetWorkItemsArgs) (*[]workitemtracking.WorkItem, error) {
	panic("GetWorkItems not implemented in mock")
}

// GetWorkItemsBatch delegates to GetWorkItemsBatchFn.
func (m *mockWITClient) GetWorkItemsBatch(ctx context.Context, args workitemtracking.GetWorkItemsBatchArgs) (*[]workitemtracking.WorkItem, error) {
	if m.GetWorkItemsBatchFn != nil {
		return m.GetWorkItemsBatchFn(ctx, args)
	}

	panic("GetWorkItemsBatchFn not set")
}

func (m *mockWITClient) GetWorkItemTemplate(context.Context, workitemtracking.GetWorkItemTemplateArgs) (*workitemtracking.WorkItem, error) {
	panic("GetWorkItemTemplate not implemented in mock")
}

func (m *mockWITClient) GetWorkItemType(context.Context, workitemtracking.GetWorkItemTypeArgs) (*workitemtracking.WorkItemType, error) {
	panic("GetWorkItemType not implemented in mock")
}

func (m *mockWITClient) GetWorkItemTypeCategories(context.Context, workitemtracking.GetWorkItemTypeCategoriesArgs) (*[]workitemtracking.WorkItemTypeCategory, error) {
	panic("GetWorkItemTypeCategories not implemented in mock")
}

func (m *mockWITClient) GetWorkItemTypeCategory(context.Context, workitemtracking.GetWorkItemTypeCategoryArgs) (*workitemtracking.WorkItemTypeCategory, error) {
	panic("GetWorkItemTypeCategory not implemented in mock")
}

func (m *mockWITClient) GetWorkItemTypeFieldsWithReferences(context.Context, workitemtracking.GetWorkItemTypeFieldsWithReferencesArgs) (*[]workitemtracking.WorkItemTypeFieldWithReferences, error) {
	panic("GetWorkItemTypeFieldsWithReferences not implemented in mock")
}

func (m *mockWITClient) GetWorkItemTypeFieldWithReferences(context.Context, workitemtracking.GetWorkItemTypeFieldWithReferencesArgs) (*workitemtracking.WorkItemTypeFieldWithReferences, error) {
	panic("GetWorkItemTypeFieldWithReferences not implemented in mock")
}

func (m *mockWITClient) GetWorkItemTypes(context.Context, workitemtracking.GetWorkItemTypesArgs) (*[]workitemtracking.WorkItemType, error) {
	panic("GetWorkItemTypes not implemented in mock")
}

func (m *mockWITClient) GetWorkItemTypeStates(context.Context, workitemtracking.GetWorkItemTypeStatesArgs) (*[]workitemtracking.WorkItemStateColor, error) {
	panic("GetWorkItemTypeStates not implemented in mock")
}

func (m *mockWITClient) MigrateProjectsProcess(context.Context, workitemtracking.MigrateProjectsProcessArgs) (*workitemtracking.ProcessMigrationResultModel, error) {
	panic("MigrateProjectsProcess not implemented in mock")
}

func (m *mockWITClient) QueryById(context.Context, workitemtracking.QueryByIdArgs) (*workitemtracking.WorkItemQueryResult, error) {
	panic("QueryById not implemented in mock")
}

func (m *mockWITClient) QueryWorkItemsForArtifactUris(context.Context, workitemtracking.QueryWorkItemsForArtifactUrisArgs) (*workitemtracking.ArtifactUriQueryResult, error) {
	panic("QueryWorkItemsForArtifactUris not implemented in mock")
}

func (m *mockWITClient) ReadReportingDiscussions(context.Context, workitemtracking.ReadReportingDiscussionsArgs) (*workitemtracking.ReportingWorkItemRevisionsBatch, error) {
	panic("ReadReportingDiscussions not implemented in mock")
}

func (m *mockWITClient) ReadReportingRevisionsGet(context.Context, workitemtracking.ReadReportingRevisionsGetArgs) (*workitemtracking.ReportingWorkItemRevisionsBatch, error) {
	panic("ReadReportingRevisionsGet not implemented in mock")
}

func (m *mockWITClient) ReadReportingRevisionsPost(context.Context, workitemtracking.ReadReportingRevisionsPostArgs) (*workitemtracking.ReportingWorkItemRevisionsBatch, error) {
	panic("ReadReportingRevisionsPost not implemented in mock")
}

func (m *mockWITClient) ReplaceTemplate(context.Context, workitemtracking.ReplaceTemplateArgs) (*workitemtracking.WorkItemTemplate, error) {
	panic("ReplaceTemplate not implemented in mock")
}

func (m *mockWITClient) RestoreWorkItem(context.Context, workitemtracking.RestoreWorkItemArgs) (*workitemtracking.WorkItemDelete, error) {
	panic("RestoreWorkItem not implemented in mock")
}

func (m *mockWITClient) SearchQueries(context.Context, workitemtracking.SearchQueriesArgs) (*workitemtracking.QueryHierarchyItemsResult, error) {
	panic("SearchQueries not implemented in mock")
}

func (m *mockWITClient) SendMail(context.Context, workitemtracking.SendMailArgs) error {
	panic("SendMail not implemented in mock")
}

func (m *mockWITClient) UpdateClassificationNode(context.Context, workitemtracking.UpdateClassificationNodeArgs) (*workitemtracking.WorkItemClassificationNode, error) {
	panic("UpdateClassificationNode not implemented in mock")
}

func (m *mockWITClient) UpdateComment(context.Context, workitemtracking.UpdateCommentArgs) (*workitemtracking.Comment, error) {
	panic("UpdateComment not implemented in mock")
}

func (m *mockWITClient) UpdateGithubConnectionRepos(context.Context, workitemtracking.UpdateGithubConnectionReposArgs) (*[]workitemtracking.GitHubConnectionRepoModel, error) {
	panic("UpdateGithubConnectionRepos not implemented in mock")
}

func (m *mockWITClient) UpdateQuery(context.Context, workitemtracking.UpdateQueryArgs) (*workitemtracking.QueryHierarchyItem, error) {
	panic("UpdateQuery not implemented in mock")
}

func (m *mockWITClient) UpdateTag(context.Context, workitemtracking.UpdateTagArgs) (*workitemtracking.WorkItemTagDefinition, error) {
	panic("UpdateTag not implemented in mock")
}

func (m *mockWITClient) UpdateWorkItemComment(context.Context, workitemtracking.UpdateWorkItemCommentArgs) (*workitemtracking.Comment, error) {
	panic("UpdateWorkItemComment not implemented in mock")
}

func (m *mockWITClient) UpdateWorkItemField(context.Context, workitemtracking.UpdateWorkItemFieldArgs) (*workitemtracking.WorkItemField2, error) {
	panic("UpdateWorkItemField not implemented in mock")
}
