/*
Daytona Server API

Daytona Server API

API version: 0.1.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package api_client

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
)


// WorkspaceAPIService WorkspaceAPI service
type WorkspaceAPIService service

type ApiCreateWorkspaceRequest struct {
	ctx context.Context
	ApiService *WorkspaceAPIService
	workspace *CreateWorkspace
}

// Create workspace
func (r ApiCreateWorkspaceRequest) Workspace(workspace CreateWorkspace) ApiCreateWorkspaceRequest {
	r.workspace = &workspace
	return r
}

func (r ApiCreateWorkspaceRequest) Execute() (*Workspace, *http.Response, error) {
	return r.ApiService.CreateWorkspaceExecute(r)
}

/*
CreateWorkspace Create a workspace

Create a workspace

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @return ApiCreateWorkspaceRequest
*/
func (a *WorkspaceAPIService) CreateWorkspace(ctx context.Context) ApiCreateWorkspaceRequest {
	return ApiCreateWorkspaceRequest{
		ApiService: a,
		ctx: ctx,
	}
}

// Execute executes the request
//  @return Workspace
func (a *WorkspaceAPIService) CreateWorkspaceExecute(r ApiCreateWorkspaceRequest) (*Workspace, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodPost
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *Workspace
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "WorkspaceAPIService.CreateWorkspace")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/workspace"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}
	if r.workspace == nil {
		return localVarReturnValue, nil, reportError("workspace is required and must be specified")
	}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	// body params
	localVarPostBody = r.workspace
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiGetWorkspaceInfoRequest struct {
	ctx context.Context
	ApiService *WorkspaceAPIService
	workspaceId string
}

func (r ApiGetWorkspaceInfoRequest) Execute() (*WorkspaceInfo, *http.Response, error) {
	return r.ApiService.GetWorkspaceInfoExecute(r)
}

/*
GetWorkspaceInfo Get workspace info

Get workspace info

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param workspaceId Workspace ID
 @return ApiGetWorkspaceInfoRequest
*/
func (a *WorkspaceAPIService) GetWorkspaceInfo(ctx context.Context, workspaceId string) ApiGetWorkspaceInfoRequest {
	return ApiGetWorkspaceInfoRequest{
		ApiService: a,
		ctx: ctx,
		workspaceId: workspaceId,
	}
}

// Execute executes the request
//  @return WorkspaceInfo
func (a *WorkspaceAPIService) GetWorkspaceInfoExecute(r ApiGetWorkspaceInfoRequest) (*WorkspaceInfo, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodGet
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *WorkspaceInfo
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "WorkspaceAPIService.GetWorkspaceInfo")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/workspace/{workspaceId}"
	localVarPath = strings.Replace(localVarPath, "{"+"workspaceId"+"}", url.PathEscape(parameterValueToString(r.workspaceId, "workspaceId")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiListWorkspacesRequest struct {
	ctx context.Context
	ApiService *WorkspaceAPIService
}

func (r ApiListWorkspacesRequest) Execute() ([]WorkspaceInfo, *http.Response, error) {
	return r.ApiService.ListWorkspacesExecute(r)
}

/*
ListWorkspaces List workspaces info

List workspaces info

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @return ApiListWorkspacesRequest
*/
func (a *WorkspaceAPIService) ListWorkspaces(ctx context.Context) ApiListWorkspacesRequest {
	return ApiListWorkspacesRequest{
		ApiService: a,
		ctx: ctx,
	}
}

// Execute executes the request
//  @return []WorkspaceInfo
func (a *WorkspaceAPIService) ListWorkspacesExecute(r ApiListWorkspacesRequest) ([]WorkspaceInfo, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodGet
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  []WorkspaceInfo
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "WorkspaceAPIService.ListWorkspaces")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/workspace"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiRemoveWorkspaceRequest struct {
	ctx context.Context
	ApiService *WorkspaceAPIService
	workspaceId string
}

func (r ApiRemoveWorkspaceRequest) Execute() (*http.Response, error) {
	return r.ApiService.RemoveWorkspaceExecute(r)
}

/*
RemoveWorkspace Remove workspace

Remove workspace

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param workspaceId Workspace ID
 @return ApiRemoveWorkspaceRequest
*/
func (a *WorkspaceAPIService) RemoveWorkspace(ctx context.Context, workspaceId string) ApiRemoveWorkspaceRequest {
	return ApiRemoveWorkspaceRequest{
		ApiService: a,
		ctx: ctx,
		workspaceId: workspaceId,
	}
}

// Execute executes the request
func (a *WorkspaceAPIService) RemoveWorkspaceExecute(r ApiRemoveWorkspaceRequest) (*http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodDelete
		localVarPostBody     interface{}
		formFiles            []formFile
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "WorkspaceAPIService.RemoveWorkspace")
	if err != nil {
		return nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/workspace/{workspaceId}"
	localVarPath = strings.Replace(localVarPath, "{"+"workspaceId"+"}", url.PathEscape(parameterValueToString(r.workspaceId, "workspaceId")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarHTTPResponse, newErr
	}

	return localVarHTTPResponse, nil
}

type ApiStartProjectRequest struct {
	ctx context.Context
	ApiService *WorkspaceAPIService
	workspaceId string
	projectId string
}

func (r ApiStartProjectRequest) Execute() (*http.Response, error) {
	return r.ApiService.StartProjectExecute(r)
}

/*
StartProject Start project

Start project

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param workspaceId Workspace ID
 @param projectId Project ID
 @return ApiStartProjectRequest
*/
func (a *WorkspaceAPIService) StartProject(ctx context.Context, workspaceId string, projectId string) ApiStartProjectRequest {
	return ApiStartProjectRequest{
		ApiService: a,
		ctx: ctx,
		workspaceId: workspaceId,
		projectId: projectId,
	}
}

// Execute executes the request
func (a *WorkspaceAPIService) StartProjectExecute(r ApiStartProjectRequest) (*http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodPost
		localVarPostBody     interface{}
		formFiles            []formFile
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "WorkspaceAPIService.StartProject")
	if err != nil {
		return nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/workspace/{workspaceId}/{projectId}/start"
	localVarPath = strings.Replace(localVarPath, "{"+"workspaceId"+"}", url.PathEscape(parameterValueToString(r.workspaceId, "workspaceId")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"projectId"+"}", url.PathEscape(parameterValueToString(r.projectId, "projectId")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarHTTPResponse, newErr
	}

	return localVarHTTPResponse, nil
}

type ApiStartWorkspaceRequest struct {
	ctx context.Context
	ApiService *WorkspaceAPIService
	workspaceId string
}

func (r ApiStartWorkspaceRequest) Execute() (*http.Response, error) {
	return r.ApiService.StartWorkspaceExecute(r)
}

/*
StartWorkspace Start workspace

Start workspace

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param workspaceId Workspace ID
 @return ApiStartWorkspaceRequest
*/
func (a *WorkspaceAPIService) StartWorkspace(ctx context.Context, workspaceId string) ApiStartWorkspaceRequest {
	return ApiStartWorkspaceRequest{
		ApiService: a,
		ctx: ctx,
		workspaceId: workspaceId,
	}
}

// Execute executes the request
func (a *WorkspaceAPIService) StartWorkspaceExecute(r ApiStartWorkspaceRequest) (*http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodPost
		localVarPostBody     interface{}
		formFiles            []formFile
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "WorkspaceAPIService.StartWorkspace")
	if err != nil {
		return nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/workspace/{workspaceId}/start"
	localVarPath = strings.Replace(localVarPath, "{"+"workspaceId"+"}", url.PathEscape(parameterValueToString(r.workspaceId, "workspaceId")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarHTTPResponse, newErr
	}

	return localVarHTTPResponse, nil
}

type ApiStopProjectRequest struct {
	ctx context.Context
	ApiService *WorkspaceAPIService
	workspaceId string
	projectId string
}

func (r ApiStopProjectRequest) Execute() (*http.Response, error) {
	return r.ApiService.StopProjectExecute(r)
}

/*
StopProject Stop project

Stop project

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param workspaceId Workspace ID
 @param projectId Project ID
 @return ApiStopProjectRequest
*/
func (a *WorkspaceAPIService) StopProject(ctx context.Context, workspaceId string, projectId string) ApiStopProjectRequest {
	return ApiStopProjectRequest{
		ApiService: a,
		ctx: ctx,
		workspaceId: workspaceId,
		projectId: projectId,
	}
}

// Execute executes the request
func (a *WorkspaceAPIService) StopProjectExecute(r ApiStopProjectRequest) (*http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodPost
		localVarPostBody     interface{}
		formFiles            []formFile
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "WorkspaceAPIService.StopProject")
	if err != nil {
		return nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/workspace/{workspaceId}/{projectId}/stop"
	localVarPath = strings.Replace(localVarPath, "{"+"workspaceId"+"}", url.PathEscape(parameterValueToString(r.workspaceId, "workspaceId")), -1)
	localVarPath = strings.Replace(localVarPath, "{"+"projectId"+"}", url.PathEscape(parameterValueToString(r.projectId, "projectId")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarHTTPResponse, newErr
	}

	return localVarHTTPResponse, nil
}

type ApiStopWorkspaceRequest struct {
	ctx context.Context
	ApiService *WorkspaceAPIService
	workspaceId string
}

func (r ApiStopWorkspaceRequest) Execute() (*http.Response, error) {
	return r.ApiService.StopWorkspaceExecute(r)
}

/*
StopWorkspace Stop workspace

Stop workspace

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param workspaceId Workspace ID
 @return ApiStopWorkspaceRequest
*/
func (a *WorkspaceAPIService) StopWorkspace(ctx context.Context, workspaceId string) ApiStopWorkspaceRequest {
	return ApiStopWorkspaceRequest{
		ApiService: a,
		ctx: ctx,
		workspaceId: workspaceId,
	}
}

// Execute executes the request
func (a *WorkspaceAPIService) StopWorkspaceExecute(r ApiStopWorkspaceRequest) (*http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodPost
		localVarPostBody     interface{}
		formFiles            []formFile
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "WorkspaceAPIService.StopWorkspace")
	if err != nil {
		return nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/workspace/{workspaceId}/stop"
	localVarPath = strings.Replace(localVarPath, "{"+"workspaceId"+"}", url.PathEscape(parameterValueToString(r.workspaceId, "workspaceId")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarHTTPResponse, err
	}

	localVarBody, err := io.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = io.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarHTTPResponse, newErr
	}

	return localVarHTTPResponse, nil
}
