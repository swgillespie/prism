// Code generated by protoc-gen-go_temporal. DO NOT EDIT.
// versions:
//
//	protoc-gen-go_temporal dev (latest)
//	go go1.21.3
//	protoc (unknown)
//
// source: prism/ingest/v1/ingest.proto
package ingestv1

import (
	"context"
	"errors"
	"fmt"
	expression "github.com/cludden/protoc-gen-go-temporal/pkg/expression"
	v2 "github.com/urfave/cli/v2"
	v1 "go.temporal.io/api/enums/v1"
	activity "go.temporal.io/sdk/activity"
	client "go.temporal.io/sdk/client"
	testsuite "go.temporal.io/sdk/testsuite"
	worker "go.temporal.io/sdk/worker"
	workflow "go.temporal.io/sdk/workflow"
	"sort"
)

// IngestTaskQueue= is the default task-queue for a prism.ingest.v1.Ingest worker
const IngestTaskQueue = "prism-ingest"

// prism.ingest.v1.Ingest workflow names
const (
	IngestObjectWorkflowName = "prism.ingest.v1.Ingest.IngestObject"
)

// prism.ingest.v1.Ingest workflow id expressions
var (
	IngestObjectIDExpression = expression.MustParseExpression("ingest/${!tenant_id.slug()}/${!table.slug()}/${!location.slug()}")
)

// prism.ingest.v1.Ingest activity names
const (
	RecordNewPartitionActivityName = "prism.ingest.v1.Ingest.RecordNewPartition"
	TransformToParquetActivityName = "prism.ingest.v1.Ingest.TransformToParquet"
)

// IngestClient describes a client for a(n) prism.ingest.v1.Ingest worker
type IngestClient interface {
	// IngestObject executes a(n) prism.ingest.v1.Ingest.IngestObject workflow and blocks until error or response received
	IngestObject(ctx context.Context, req *IngestObjectRequest, opts ...*IngestObjectOptions) error
	// IngestObjectAsync executes a(n) prism.ingest.v1.Ingest.IngestObject workflow asynchronously
	IngestObjectAsync(ctx context.Context, req *IngestObjectRequest, opts ...*IngestObjectOptions) (IngestObjectRun, error)
	// GetIngestObject retrieves a handle to an existing prism.ingest.v1.Ingest.IngestObject workflow execution
	GetIngestObject(ctx context.Context, workflowID string, runID string) IngestObjectRun
	// CancelWorkflow requests cancellation of an existing workflow execution
	CancelWorkflow(ctx context.Context, workflowID string, runID string) error
	// TerminateWorkflow an existing workflow execution
	TerminateWorkflow(ctx context.Context, workflowID string, runID string, reason string, details ...interface{}) error
}

// ingestClient implements a temporal client for a prism.ingest.v1.Ingest service
type ingestClient struct {
	client client.Client
}

// NewIngestClient initializes a new prism.ingest.v1.Ingest client
func NewIngestClient(c client.Client) IngestClient {
	return &ingestClient{client: c}
}

// NewIngestClientWithOptions initializes a new Ingest client with the given options
func NewIngestClientWithOptions(c client.Client, opts client.Options) (IngestClient, error) {
	var err error
	c, err = client.NewClientFromExisting(c, opts)
	if err != nil {
		return nil, fmt.Errorf("error initializing client with options: %w", err)
	}
	return &ingestClient{client: c}, nil
}

// IngestObject executes a prism.ingest.v1.Ingest.IngestObject workflow and blocks until error or response received
func (c *ingestClient) IngestObject(ctx context.Context, req *IngestObjectRequest, options ...*IngestObjectOptions) error {
	run, err := c.IngestObjectAsync(ctx, req, options...)
	if err != nil {
		return err
	}
	return run.Get(ctx)
}

// IngestObjectAsync starts a(n) prism.ingest.v1.Ingest.IngestObject workflow
func (c *ingestClient) IngestObjectAsync(ctx context.Context, req *IngestObjectRequest, options ...*IngestObjectOptions) (IngestObjectRun, error) {
	opts := &client.StartWorkflowOptions{}
	if len(options) > 0 && options[0].opts != nil {
		opts = options[0].opts
	}
	if opts.TaskQueue == "" {
		opts.TaskQueue = IngestTaskQueue
	}
	if opts.ID == "" {
		id, err := expression.EvalExpression(IngestObjectIDExpression, req.ProtoReflect())
		if err != nil {
			return nil, fmt.Errorf("error evaluating id expression for \"IngestObject\" workflow: %w", err)
		}
		opts.ID = id
	}
	if opts.WorkflowIDReusePolicy == v1.WORKFLOW_ID_REUSE_POLICY_UNSPECIFIED {
		opts.WorkflowIDReusePolicy = v1.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE
	}
	if opts.WorkflowExecutionTimeout == 0 {
		opts.WorkflowExecutionTimeout = 360000000000 // 6m0s
	}
	run, err := c.client.ExecuteWorkflow(ctx, *opts, IngestObjectWorkflowName, req)
	if err != nil {
		return nil, err
	}
	if run == nil {
		return nil, errors.New("execute workflow returned nil run")
	}
	return &ingestObjectRun{
		client: c,
		run:    run,
	}, nil
}

// GetIngestObject fetches an existing prism.ingest.v1.Ingest.IngestObject execution
func (c *ingestClient) GetIngestObject(ctx context.Context, workflowID string, runID string) IngestObjectRun {
	return &ingestObjectRun{
		client: c,
		run:    c.client.GetWorkflow(ctx, workflowID, runID),
	}
}

// CancelWorkflow requests cancellation of an existing workflow execution
func (c *ingestClient) CancelWorkflow(ctx context.Context, workflowID string, runID string) error {
	return c.client.CancelWorkflow(ctx, workflowID, runID)
}

// TerminateWorkflow terminates an existing workflow execution
func (c *ingestClient) TerminateWorkflow(ctx context.Context, workflowID string, runID string, reason string, details ...interface{}) error {
	return c.client.TerminateWorkflow(ctx, workflowID, runID, reason, details...)
}

// IngestObjectOptions provides configuration for a prism.ingest.v1.Ingest.IngestObject workflow operation
type IngestObjectOptions struct {
	opts *client.StartWorkflowOptions
}

// NewIngestObjectOptions initializes a new IngestObjectOptions value
func NewIngestObjectOptions() *IngestObjectOptions {
	return &IngestObjectOptions{}
}

// WithStartWorkflowOptions sets the initial client.StartWorkflowOptions
func (opts *IngestObjectOptions) WithStartWorkflowOptions(options client.StartWorkflowOptions) *IngestObjectOptions {
	opts.opts = &options
	return opts
}

// IngestObjectRun describes a(n) prism.ingest.v1.Ingest.IngestObject workflow run
type IngestObjectRun interface {
	// ID returns the workflow ID
	ID() string
	// RunID returns the workflow instance ID
	RunID() string
	// Get blocks until the workflow is complete and returns the result
	Get(ctx context.Context) error
	// Cancel requests cancellation of a workflow in execution, returning an error if applicable
	Cancel(ctx context.Context) error
	// Terminate terminates a workflow in execution, returning an error if applicable
	Terminate(ctx context.Context, reason string, details ...interface{}) error
}

// ingestObjectRun provides an internal implementation of a(n) IngestObjectRunRun
type ingestObjectRun struct {
	client *ingestClient
	run    client.WorkflowRun
}

// ID returns the workflow ID
func (r *ingestObjectRun) ID() string {
	return r.run.GetID()
}

// RunID returns the execution ID
func (r *ingestObjectRun) RunID() string {
	return r.run.GetRunID()
}

// Cancel requests cancellation of a workflow in execution, returning an error if applicable
func (r *ingestObjectRun) Cancel(ctx context.Context) error {
	return r.client.CancelWorkflow(ctx, r.ID(), r.RunID())
}

// Get blocks until the workflow is complete, returning the result if applicable
func (r *ingestObjectRun) Get(ctx context.Context) error {
	return r.run.Get(ctx, nil)
}

// Terminate terminates a workflow in execution, returning an error if applicable
func (r *ingestObjectRun) Terminate(ctx context.Context, reason string, details ...interface{}) error {
	return r.client.TerminateWorkflow(ctx, r.ID(), r.RunID(), reason, details...)
}

// Reference to generated workflow functions
var (
	// IngestObjectFunction implements a "IngestObjectWorkflow" workflow
	IngestObjectFunction func(workflow.Context, *IngestObjectRequest) error
)

// IngestWorkflows provides methods for initializing new prism.ingest.v1.Ingest workflow values
type IngestWorkflows interface {
	IngestObject(ctx workflow.Context, input *IngestObjectInput) (IngestObjectWorkflow, error)
}

// IngestObject initializes a new a(n) IngestObjectWorkflow implementation
// RegisterIngestWorkflows registers prism.ingest.v1.Ingest workflows with the given worker
func RegisterIngestWorkflows(r worker.WorkflowRegistry, workflows IngestWorkflows) {
	RegisterIngestObjectWorkflow(r, workflows.IngestObject)
}

// RegisterIngestObjectWorkflow registers a prism.ingest.v1.Ingest.IngestObject workflow with the given worker
func RegisterIngestObjectWorkflow(r worker.WorkflowRegistry, wf func(workflow.Context, *IngestObjectInput) (IngestObjectWorkflow, error)) {
	IngestObjectFunction = buildIngestObject(wf)
	r.RegisterWorkflowWithOptions(IngestObjectFunction, workflow.RegisterOptions{Name: IngestObjectWorkflowName})
}

// buildIngestObject converts a IngestObject workflow struct into a valid workflow function
func buildIngestObject(ctor func(workflow.Context, *IngestObjectInput) (IngestObjectWorkflow, error)) func(workflow.Context, *IngestObjectRequest) error {
	return func(ctx workflow.Context, req *IngestObjectRequest) error {
		input := &IngestObjectInput{
			Req: req,
		}
		wf, err := ctor(ctx, input)
		if err != nil {
			return err
		}
		return wf.Execute(ctx)
	}
}

// IngestObjectInput describes the input to a(n) prism.ingest.v1.Ingest.IngestObject workflow constructor
type IngestObjectInput struct {
	Req *IngestObjectRequest
}

// IngestObjectWorkflow describes a(n) prism.ingest.v1.Ingest.IngestObject workflow implementation
type IngestObjectWorkflow interface {
	// Execute defines the entrypoint to a(n) prism.ingest.v1.Ingest.IngestObject workflow
	Execute(ctx workflow.Context) error
}

// IngestObjectChild executes a child prism.ingest.v1.Ingest.IngestObject workflow
func IngestObjectChild(ctx workflow.Context, req *IngestObjectRequest, options ...*IngestObjectChildOptions) error {
	childRun, err := IngestObjectChildAsync(ctx, req, options...)
	if err != nil {
		return err
	}
	return childRun.Get(ctx)
}

// IngestObjectChildAsync executes a child prism.ingest.v1.Ingest.IngestObject workflow
func IngestObjectChildAsync(ctx workflow.Context, req *IngestObjectRequest, options ...*IngestObjectChildOptions) (*IngestObjectChildRun, error) {
	var opts *workflow.ChildWorkflowOptions
	if len(options) > 0 && options[0].opts != nil {
		opts = options[0].opts
	} else {
		childOpts := workflow.GetChildWorkflowOptions(ctx)
		opts = &childOpts
	}
	if opts.TaskQueue == "" {
		opts.TaskQueue = IngestTaskQueue
	}
	if opts.WorkflowID == "" {
		id, err := expression.EvalExpression(IngestObjectIDExpression, req.ProtoReflect())
		if err != nil {
			panic(err)
		}
		opts.WorkflowID = id
	}
	if opts.WorkflowIDReusePolicy == v1.WORKFLOW_ID_REUSE_POLICY_UNSPECIFIED {
		opts.WorkflowIDReusePolicy = v1.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE
	}
	if opts.WorkflowExecutionTimeout == 0 {
		opts.WorkflowExecutionTimeout = 360000000000 // 6m0s
	}
	ctx = workflow.WithChildOptions(ctx, *opts)
	return &IngestObjectChildRun{Future: workflow.ExecuteChildWorkflow(ctx, IngestObjectWorkflowName, req)}, nil
}

// IngestObjectChildOptions provides configuration for a prism.ingest.v1.Ingest.IngestObject workflow operation
type IngestObjectChildOptions struct {
	opts *workflow.ChildWorkflowOptions
}

// NewIngestObjectChildOptions initializes a new IngestObjectChildOptions value
func NewIngestObjectChildOptions() *IngestObjectChildOptions {
	return &IngestObjectChildOptions{}
}

// WithChildWorkflowOptions sets the initial client.StartWorkflowOptions
func (opts *IngestObjectChildOptions) WithChildWorkflowOptions(options workflow.ChildWorkflowOptions) *IngestObjectChildOptions {
	opts.opts = &options
	return opts
}

// IngestObjectChildRun describes a child prism.ingest.v1.Ingest.IngestObject workflow run
type IngestObjectChildRun struct {
	Future workflow.ChildWorkflowFuture
}

// Get blocks until the workflow is completed, returning the response value
func (r *IngestObjectChildRun) Get(ctx workflow.Context) error {
	if err := r.Future.Get(ctx, nil); err != nil {
		return err
	}
	return nil
}

// Select adds this completion to the selector. Callback can be nil.
func (r *IngestObjectChildRun) Select(sel workflow.Selector, fn func(*IngestObjectChildRun)) workflow.Selector {
	return sel.AddFuture(r.Future, func(workflow.Future) {
		if fn != nil {
			fn(r)
		}
	})
}

// SelectStart adds waiting for start to the selector. Callback can be nil.
func (r *IngestObjectChildRun) SelectStart(sel workflow.Selector, fn func(*IngestObjectChildRun)) workflow.Selector {
	return sel.AddFuture(r.Future.GetChildWorkflowExecution(), func(workflow.Future) {
		if fn != nil {
			fn(r)
		}
	})
}

// WaitStart waits for the child workflow to start
func (r *IngestObjectChildRun) WaitStart(ctx workflow.Context) (*workflow.Execution, error) {
	var exec workflow.Execution
	if err := r.Future.GetChildWorkflowExecution().Get(ctx, &exec); err != nil {
		return nil, err
	}
	return &exec, nil
}

// IngestActivities describes available worker activites
type IngestActivities interface {
	RecordNewPartition(ctx context.Context, req *RecordNewPartitionRequest) error
	TransformToParquet(ctx context.Context, req *TransformToParquetRequest) (*TransformToParquetResponse, error)
}

// RegisterIngestActivities registers activities with a worker
func RegisterIngestActivities(r worker.ActivityRegistry, activities IngestActivities) {
	RegisterRecordNewPartitionActivity(r, activities.RecordNewPartition)
	RegisterTransformToParquetActivity(r, activities.TransformToParquet)
}

// RegisterRecordNewPartitionActivity registers a prism.ingest.v1.Ingest.RecordNewPartition activity
func RegisterRecordNewPartitionActivity(r worker.ActivityRegistry, fn func(context.Context, *RecordNewPartitionRequest) error) {
	r.RegisterActivityWithOptions(fn, activity.RegisterOptions{
		Name: RecordNewPartitionActivityName,
	})
}

// RecordNewPartitionFuture describes a(n) prism.ingest.v1.Ingest.RecordNewPartition activity execution
type RecordNewPartitionFuture struct {
	Future workflow.Future
}

// Get blocks on the activity's completion, returning the response
func (f *RecordNewPartitionFuture) Get(ctx workflow.Context) error {
	return f.Future.Get(ctx, nil)
}

// Select adds the activity's completion to the selector, callback can be nil
func (f *RecordNewPartitionFuture) Select(sel workflow.Selector, fn func(*RecordNewPartitionFuture)) workflow.Selector {
	return sel.AddFuture(f.Future, func(workflow.Future) {
		if fn != nil {
			fn(f)
		}
	})
}

// RecordNewPartition executes a(n) prism.ingest.v1.Ingest.RecordNewPartition activity
func RecordNewPartition(ctx workflow.Context, req *RecordNewPartitionRequest, options ...*RecordNewPartitionActivityOptions) error {
	var opts *RecordNewPartitionActivityOptions
	if len(options) > 0 && options[0] != nil {
		opts = options[0]
	} else {
		opts = NewRecordNewPartitionActivityOptions()
	}
	if opts.opts == nil {
		activityOpts := workflow.GetActivityOptions(ctx)
		opts.opts = &activityOpts
	}
	if opts.opts.StartToCloseTimeout == 0 {
		opts.opts.StartToCloseTimeout = 30000000000 // 30s
	}
	ctx = workflow.WithActivityOptions(ctx, *opts.opts)
	var activity any
	activity = RecordNewPartitionActivityName
	future := &RecordNewPartitionFuture{Future: workflow.ExecuteActivity(ctx, activity, req)}
	return future.Get(ctx)
}

// RecordNewPartitionAsync executes a(n) prism.ingest.v1.Ingest.RecordNewPartition activity (asynchronously)
func RecordNewPartitionAsync(ctx workflow.Context, req *RecordNewPartitionRequest, options ...*RecordNewPartitionActivityOptions) *RecordNewPartitionFuture {
	var opts *RecordNewPartitionActivityOptions
	if len(options) > 0 && options[0] != nil {
		opts = options[0]
	} else {
		opts = NewRecordNewPartitionActivityOptions()
	}
	if opts.opts == nil {
		activityOpts := workflow.GetActivityOptions(ctx)
		opts.opts = &activityOpts
	}
	if opts.opts.StartToCloseTimeout == 0 {
		opts.opts.StartToCloseTimeout = 30000000000 // 30s
	}
	ctx = workflow.WithActivityOptions(ctx, *opts.opts)
	var activity any
	activity = RecordNewPartitionActivityName
	future := &RecordNewPartitionFuture{Future: workflow.ExecuteActivity(ctx, activity, req)}
	return future
}

// RecordNewPartitionLocal executes a(n) prism.ingest.v1.Ingest.RecordNewPartition activity (locally)
func RecordNewPartitionLocal(ctx workflow.Context, req *RecordNewPartitionRequest, options ...*RecordNewPartitionLocalActivityOptions) error {
	var opts *RecordNewPartitionLocalActivityOptions
	if len(options) > 0 && options[0] != nil {
		opts = options[0]
	} else {
		opts = NewRecordNewPartitionLocalActivityOptions()
	}
	if opts.opts == nil {
		activityOpts := workflow.GetLocalActivityOptions(ctx)
		opts.opts = &activityOpts
	}
	if opts.opts.StartToCloseTimeout == 0 {
		opts.opts.StartToCloseTimeout = 30000000000 // 30s
	}
	ctx = workflow.WithLocalActivityOptions(ctx, *opts.opts)
	var activity any
	if opts.fn != nil {
		activity = opts.fn
	} else {
		activity = RecordNewPartitionActivityName
	}
	future := &RecordNewPartitionFuture{Future: workflow.ExecuteLocalActivity(ctx, activity, req)}
	return future.Get(ctx)
}

// RecordNewPartitionLocalAsync executes a(n) prism.ingest.v1.Ingest.RecordNewPartition activity (asynchronously, locally)
func RecordNewPartitionLocalAsync(ctx workflow.Context, req *RecordNewPartitionRequest, options ...*RecordNewPartitionLocalActivityOptions) *RecordNewPartitionFuture {
	var opts *RecordNewPartitionLocalActivityOptions
	if len(options) > 0 && options[0] != nil {
		opts = options[0]
	} else {
		opts = NewRecordNewPartitionLocalActivityOptions()
	}
	if opts.opts == nil {
		activityOpts := workflow.GetLocalActivityOptions(ctx)
		opts.opts = &activityOpts
	}
	if opts.opts.StartToCloseTimeout == 0 {
		opts.opts.StartToCloseTimeout = 30000000000 // 30s
	}
	ctx = workflow.WithLocalActivityOptions(ctx, *opts.opts)
	var activity any
	if opts.fn != nil {
		activity = opts.fn
	} else {
		activity = RecordNewPartitionActivityName
	}
	future := &RecordNewPartitionFuture{Future: workflow.ExecuteLocalActivity(ctx, activity, req)}
	return future
}

// RecordNewPartitionLocalActivityOptions provides configuration for a local prism.ingest.v1.Ingest.RecordNewPartition activity
type RecordNewPartitionLocalActivityOptions struct {
	fn   func(context.Context, *RecordNewPartitionRequest) error
	opts *workflow.LocalActivityOptions
}

// NewRecordNewPartitionLocalActivityOptions sets default LocalActivityOptions
func NewRecordNewPartitionLocalActivityOptions() *RecordNewPartitionLocalActivityOptions {
	return &RecordNewPartitionLocalActivityOptions{}
}

// Local provides a local prism.ingest.v1.Ingest.RecordNewPartition activity implementation
func (opts *RecordNewPartitionLocalActivityOptions) Local(fn func(context.Context, *RecordNewPartitionRequest) error) *RecordNewPartitionLocalActivityOptions {
	opts.fn = fn
	return opts
}

// WithLocalActivityOptions sets default LocalActivityOptions
func (opts *RecordNewPartitionLocalActivityOptions) WithLocalActivityOptions(options workflow.LocalActivityOptions) *RecordNewPartitionLocalActivityOptions {
	opts.opts = &options
	return opts
}

// RecordNewPartitionActivityOptions provides configuration for a(n) prism.ingest.v1.Ingest.RecordNewPartition activity
type RecordNewPartitionActivityOptions struct {
	opts *workflow.ActivityOptions
}

// NewRecordNewPartitionActivityOptions sets default ActivityOptions
func NewRecordNewPartitionActivityOptions() *RecordNewPartitionActivityOptions {
	return &RecordNewPartitionActivityOptions{}
}

// WithActivityOptions sets default ActivityOptions
func (opts *RecordNewPartitionActivityOptions) WithActivityOptions(options workflow.ActivityOptions) *RecordNewPartitionActivityOptions {
	opts.opts = &options
	return opts
}

// RegisterTransformToParquetActivity registers a prism.ingest.v1.Ingest.TransformToParquet activity
func RegisterTransformToParquetActivity(r worker.ActivityRegistry, fn func(context.Context, *TransformToParquetRequest) (*TransformToParquetResponse, error)) {
	r.RegisterActivityWithOptions(fn, activity.RegisterOptions{
		Name: TransformToParquetActivityName,
	})
}

// TransformToParquetFuture describes a(n) prism.ingest.v1.Ingest.TransformToParquet activity execution
type TransformToParquetFuture struct {
	Future workflow.Future
}

// Get blocks on the activity's completion, returning the response
func (f *TransformToParquetFuture) Get(ctx workflow.Context) (*TransformToParquetResponse, error) {
	var resp TransformToParquetResponse
	if err := f.Future.Get(ctx, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Select adds the activity's completion to the selector, callback can be nil
func (f *TransformToParquetFuture) Select(sel workflow.Selector, fn func(*TransformToParquetFuture)) workflow.Selector {
	return sel.AddFuture(f.Future, func(workflow.Future) {
		if fn != nil {
			fn(f)
		}
	})
}

// TransformToParquet executes a(n) prism.ingest.v1.Ingest.TransformToParquet activity
func TransformToParquet(ctx workflow.Context, req *TransformToParquetRequest, options ...*TransformToParquetActivityOptions) (*TransformToParquetResponse, error) {
	var opts *TransformToParquetActivityOptions
	if len(options) > 0 && options[0] != nil {
		opts = options[0]
	} else {
		opts = NewTransformToParquetActivityOptions()
	}
	if opts.opts == nil {
		activityOpts := workflow.GetActivityOptions(ctx)
		opts.opts = &activityOpts
	}
	if opts.opts.StartToCloseTimeout == 0 {
		opts.opts.StartToCloseTimeout = 360000000000 // 6m0s
	}
	ctx = workflow.WithActivityOptions(ctx, *opts.opts)
	var activity any
	activity = TransformToParquetActivityName
	future := &TransformToParquetFuture{Future: workflow.ExecuteActivity(ctx, activity, req)}
	return future.Get(ctx)
}

// TransformToParquetAsync executes a(n) prism.ingest.v1.Ingest.TransformToParquet activity (asynchronously)
func TransformToParquetAsync(ctx workflow.Context, req *TransformToParquetRequest, options ...*TransformToParquetActivityOptions) *TransformToParquetFuture {
	var opts *TransformToParquetActivityOptions
	if len(options) > 0 && options[0] != nil {
		opts = options[0]
	} else {
		opts = NewTransformToParquetActivityOptions()
	}
	if opts.opts == nil {
		activityOpts := workflow.GetActivityOptions(ctx)
		opts.opts = &activityOpts
	}
	if opts.opts.StartToCloseTimeout == 0 {
		opts.opts.StartToCloseTimeout = 360000000000 // 6m0s
	}
	ctx = workflow.WithActivityOptions(ctx, *opts.opts)
	var activity any
	activity = TransformToParquetActivityName
	future := &TransformToParquetFuture{Future: workflow.ExecuteActivity(ctx, activity, req)}
	return future
}

// TransformToParquetLocal executes a(n) prism.ingest.v1.Ingest.TransformToParquet activity (locally)
func TransformToParquetLocal(ctx workflow.Context, req *TransformToParquetRequest, options ...*TransformToParquetLocalActivityOptions) (*TransformToParquetResponse, error) {
	var opts *TransformToParquetLocalActivityOptions
	if len(options) > 0 && options[0] != nil {
		opts = options[0]
	} else {
		opts = NewTransformToParquetLocalActivityOptions()
	}
	if opts.opts == nil {
		activityOpts := workflow.GetLocalActivityOptions(ctx)
		opts.opts = &activityOpts
	}
	if opts.opts.StartToCloseTimeout == 0 {
		opts.opts.StartToCloseTimeout = 360000000000 // 6m0s
	}
	ctx = workflow.WithLocalActivityOptions(ctx, *opts.opts)
	var activity any
	if opts.fn != nil {
		activity = opts.fn
	} else {
		activity = TransformToParquetActivityName
	}
	future := &TransformToParquetFuture{Future: workflow.ExecuteLocalActivity(ctx, activity, req)}
	return future.Get(ctx)
}

// TransformToParquetLocalAsync executes a(n) prism.ingest.v1.Ingest.TransformToParquet activity (asynchronously, locally)
func TransformToParquetLocalAsync(ctx workflow.Context, req *TransformToParquetRequest, options ...*TransformToParquetLocalActivityOptions) *TransformToParquetFuture {
	var opts *TransformToParquetLocalActivityOptions
	if len(options) > 0 && options[0] != nil {
		opts = options[0]
	} else {
		opts = NewTransformToParquetLocalActivityOptions()
	}
	if opts.opts == nil {
		activityOpts := workflow.GetLocalActivityOptions(ctx)
		opts.opts = &activityOpts
	}
	if opts.opts.StartToCloseTimeout == 0 {
		opts.opts.StartToCloseTimeout = 360000000000 // 6m0s
	}
	ctx = workflow.WithLocalActivityOptions(ctx, *opts.opts)
	var activity any
	if opts.fn != nil {
		activity = opts.fn
	} else {
		activity = TransformToParquetActivityName
	}
	future := &TransformToParquetFuture{Future: workflow.ExecuteLocalActivity(ctx, activity, req)}
	return future
}

// TransformToParquetLocalActivityOptions provides configuration for a local prism.ingest.v1.Ingest.TransformToParquet activity
type TransformToParquetLocalActivityOptions struct {
	fn   func(context.Context, *TransformToParquetRequest) (*TransformToParquetResponse, error)
	opts *workflow.LocalActivityOptions
}

// NewTransformToParquetLocalActivityOptions sets default LocalActivityOptions
func NewTransformToParquetLocalActivityOptions() *TransformToParquetLocalActivityOptions {
	return &TransformToParquetLocalActivityOptions{}
}

// Local provides a local prism.ingest.v1.Ingest.TransformToParquet activity implementation
func (opts *TransformToParquetLocalActivityOptions) Local(fn func(context.Context, *TransformToParquetRequest) (*TransformToParquetResponse, error)) *TransformToParquetLocalActivityOptions {
	opts.fn = fn
	return opts
}

// WithLocalActivityOptions sets default LocalActivityOptions
func (opts *TransformToParquetLocalActivityOptions) WithLocalActivityOptions(options workflow.LocalActivityOptions) *TransformToParquetLocalActivityOptions {
	opts.opts = &options
	return opts
}

// TransformToParquetActivityOptions provides configuration for a(n) prism.ingest.v1.Ingest.TransformToParquet activity
type TransformToParquetActivityOptions struct {
	opts *workflow.ActivityOptions
}

// NewTransformToParquetActivityOptions sets default ActivityOptions
func NewTransformToParquetActivityOptions() *TransformToParquetActivityOptions {
	return &TransformToParquetActivityOptions{}
}

// WithActivityOptions sets default ActivityOptions
func (opts *TransformToParquetActivityOptions) WithActivityOptions(options workflow.ActivityOptions) *TransformToParquetActivityOptions {
	opts.opts = &options
	return opts
}

// TestClient provides a testsuite-compatible Client
type TestIngestClient struct {
	env       *testsuite.TestWorkflowEnvironment
	workflows IngestWorkflows
}

var _ IngestClient = &TestIngestClient{}

// NewTestIngestClient initializes a new TestIngestClient value
func NewTestIngestClient(env *testsuite.TestWorkflowEnvironment, workflows IngestWorkflows, activities IngestActivities) *TestIngestClient {
	RegisterIngestWorkflows(env, workflows)
	if activities != nil {
		RegisterIngestActivities(env, activities)
	}
	return &TestIngestClient{env, workflows}
}

// IngestObject executes a(n) IngestObject workflow in the test environment
func (c *TestIngestClient) IngestObject(ctx context.Context, req *IngestObjectRequest, opts ...*IngestObjectOptions) error {
	run, err := c.IngestObjectAsync(ctx, req, opts...)
	if err != nil {
		return err
	}
	return run.Get(ctx)
}

// IngestObjectAsync executes a(n) IngestObject workflow in the test environment
func (c *TestIngestClient) IngestObjectAsync(ctx context.Context, req *IngestObjectRequest, options ...*IngestObjectOptions) (IngestObjectRun, error) {
	opts := &client.StartWorkflowOptions{}
	if len(options) > 0 && options[0].opts != nil {
		opts = options[0].opts
	}
	if opts.TaskQueue == "" {
		opts.TaskQueue = IngestTaskQueue
	}
	if opts.ID == "" {
		id, err := expression.EvalExpression(IngestObjectIDExpression, req.ProtoReflect())
		if err != nil {
			return nil, fmt.Errorf("error evaluating id expression for \"IngestObject\" workflow: %w", err)
		}
		opts.ID = id
	}
	if opts.WorkflowIDReusePolicy == v1.WORKFLOW_ID_REUSE_POLICY_UNSPECIFIED {
		opts.WorkflowIDReusePolicy = v1.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE
	}
	if opts.WorkflowExecutionTimeout == 0 {
		opts.WorkflowExecutionTimeout = 360000000000 // 6m0s
	}
	return &testIngestObjectRun{client: c, env: c.env, opts: opts, req: req, workflows: c.workflows}, nil
}

// GetIngestObject is a noop
func (c *TestIngestClient) GetIngestObject(ctx context.Context, workflowID string, runID string) IngestObjectRun {
	return &testIngestObjectRun{env: c.env, workflows: c.workflows}
}

// CancelWorkflow requests cancellation of an existing workflow execution
func (c *TestIngestClient) CancelWorkflow(ctx context.Context, workflowID string, runID string) error {
	c.env.CancelWorkflow()
	return nil
}

// TerminateWorkflow terminates an existing workflow execution
func (c *TestIngestClient) TerminateWorkflow(ctx context.Context, workflowID string, runID string, reason string, details ...interface{}) error {
	return c.CancelWorkflow(ctx, workflowID, runID)
}

var _ IngestObjectRun = &testIngestObjectRun{}

// testIngestObjectRun provides convenience methods for interacting with a(n) IngestObject workflow in the test environment
type testIngestObjectRun struct {
	client    *TestIngestClient
	env       *testsuite.TestWorkflowEnvironment
	opts      *client.StartWorkflowOptions
	req       *IngestObjectRequest
	workflows IngestWorkflows
}

// Cancel requests cancellation of a workflow in execution, returning an error if applicable
func (r *testIngestObjectRun) Cancel(ctx context.Context) error {
	return r.client.CancelWorkflow(ctx, r.ID(), r.RunID())
}

// Get retrieves a test IngestObject workflow result
func (r *testIngestObjectRun) Get(context.Context) error {
	r.env.ExecuteWorkflow(IngestObjectWorkflowName, r.req)
	if !r.env.IsWorkflowCompleted() {
		return errors.New("workflow in progress")
	}
	if err := r.env.GetWorkflowError(); err != nil {
		return err
	}
	return nil
}

// ID returns a test IngestObject workflow run's workflow ID
func (r *testIngestObjectRun) ID() string {
	if r.opts != nil {
		return r.opts.ID
	}
	return ""
}

// RunID noop implementation
func (r *testIngestObjectRun) RunID() string {
	return ""
}

// Terminate terminates a workflow in execution, returning an error if applicable
func (r *testIngestObjectRun) Terminate(ctx context.Context, reason string, details ...interface{}) error {
	return r.client.TerminateWorkflow(ctx, r.ID(), r.RunID(), reason, details...)
}

// IngestCliOptions describes runtime configuration for prism.ingest.v1.Ingest cli
type IngestCliOptions struct {
	after            func(*v2.Context) error
	before           func(*v2.Context) error
	clientForCommand func(*v2.Context) (client.Client, error)
	worker           func(*v2.Context, client.Client) (worker.Worker, error)
}

// NewIngestCliOptions initializes a new IngestCliOptions value
func NewIngestCliOptions() *IngestCliOptions {
	return &IngestCliOptions{}
}

// WithAfter injects a custom After hook to be run after any command invocation
func (opts *IngestCliOptions) WithAfter(fn func(*v2.Context) error) *IngestCliOptions {
	opts.after = fn
	return opts
}

// WithBefore injects a custom Before hook to be run prior to any command invocation
func (opts *IngestCliOptions) WithBefore(fn func(*v2.Context) error) *IngestCliOptions {
	opts.before = fn
	return opts
}

// WithClient provides a Temporal client factory for use by commands
func (opts *IngestCliOptions) WithClient(fn func(*v2.Context) (client.Client, error)) *IngestCliOptions {
	opts.clientForCommand = fn
	return opts
}

// WithWorker provides an method for initializing a worker
func (opts *IngestCliOptions) WithWorker(fn func(*v2.Context, client.Client) (worker.Worker, error)) *IngestCliOptions {
	opts.worker = fn
	return opts
}

// NewIngestCli initializes a cli for a(n) prism.ingest.v1.Ingest service
func NewIngestCli(options ...*IngestCliOptions) (*v2.App, error) {
	commands, err := newIngestCommands(options...)
	if err != nil {
		return nil, fmt.Errorf("error initializing subcommands: %w", err)
	}
	return &v2.App{
		Name:     "ingest",
		Commands: commands,
	}, nil
}

// NewIngestCliCommand initializes a cli command for a prism.ingest.v1.Ingest service with subcommands for each query, signal, update, and workflow
func NewIngestCliCommand(options ...*IngestCliOptions) (*v2.Command, error) {
	subcommands, err := newIngestCommands(options...)
	if err != nil {
		return nil, fmt.Errorf("error initializing subcommands: %w", err)
	}
	return &v2.Command{
		Name:        "ingest",
		Subcommands: subcommands,
	}, nil
}

// newIngestCommands initializes (sub)commands for a prism.ingest.v1.Ingest cli or command
func newIngestCommands(options ...*IngestCliOptions) ([]*v2.Command, error) {
	opts := &IngestCliOptions{}
	if len(options) > 0 {
		opts = options[0]
	}
	if opts.clientForCommand == nil {
		opts.clientForCommand = func(*v2.Context) (client.Client, error) {
			return client.Dial(client.Options{})
		}
	}
	commands := []*v2.Command{
		// IngestObject executes a(n) IngestObject workflow,
		{
			Name:                   "ingest-object",
			Usage:                  "IngestObject executes a(n) IngestObject workflow",
			Category:               "WORKFLOWS",
			UseShortOptionHandling: true,
			Before:                 opts.before,
			After:                  opts.after,
			Flags: []v2.Flag{
				&v2.BoolFlag{
					Name:    "detach",
					Usage:   "run workflow in the background and print workflow and execution id",
					Aliases: []string{"d"},
				},
				&v2.StringFlag{
					Name:     "tenant-id",
					Usage:    "set the value of the operation's \"TenantId\" parameter",
					Category: "INPUT",
				},
				&v2.StringFlag{
					Name:     "table",
					Usage:    "set the value of the operation's \"Table\" parameter",
					Category: "INPUT",
				},
				&v2.StringFlag{
					Name:     "source",
					Usage:    "set the value of the operation's \"Source\" parameter",
					Category: "INPUT",
				},
				&v2.StringFlag{
					Name:     "destination",
					Usage:    "set the value of the operation's \"Destination\" parameter",
					Category: "INPUT",
				},
				&v2.StringFlag{
					Name:     "location",
					Usage:    "set the value of the operation's \"Location\" parameter",
					Category: "INPUT",
				},
			},
			Action: func(cmd *v2.Context) error {
				c, err := opts.clientForCommand(cmd)
				if err != nil {
					return fmt.Errorf("error initializing client for command: %w", err)
				}
				defer c.Close()
				client := NewIngestClient(c)
				req, err := unmarshalCliFlagsToIngestObjectRequest(cmd)
				if err != nil {
					return fmt.Errorf("error unmarshalling request: %w", err)
				}
				run, err := client.IngestObjectAsync(cmd.Context, req)
				if err != nil {
					return fmt.Errorf("error starting %s workflow: %w", IngestObjectWorkflowName, err)
				}
				if cmd.Bool("detach") {
					fmt.Println("success")
					fmt.Printf("workflow id: %s\n", run.ID())
					fmt.Printf("run id: %s\n", run.RunID())
					return nil
				}
				if err := run.Get(cmd.Context); err != nil {
					return err
				} else {
					return nil
				}
			},
		},
	}
	if opts.worker != nil {
		commands = append(commands, []*v2.Command{
			{
				Name:                   "worker",
				Usage:                  "runs a prism.ingest.v1.Ingest worker process",
				UseShortOptionHandling: true,
				Before:                 opts.before,
				After:                  opts.after,
				Action: func(cmd *v2.Context) error {
					c, err := opts.clientForCommand(cmd)
					if err != nil {
						return fmt.Errorf("error initializing client for command: %w", err)
					}
					defer c.Close()
					w, err := opts.worker(cmd, c)
					if opts.worker != nil {
						if err != nil {
							return fmt.Errorf("error initializing worker: %w", err)
						}
					}
					if err := w.Start(); err != nil {
						return fmt.Errorf("error starting worker: %w", err)
					}
					defer w.Stop()
					<-cmd.Context.Done()
					return nil
				},
			},
		}...)
	}
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name < commands[j].Name
	})
	return commands, nil
}

// unmarshalCliFlagsToIngestObjectRequest unmarshals a IngestObjectRequest from command line flags
func unmarshalCliFlagsToIngestObjectRequest(cmd *v2.Context) (*IngestObjectRequest, error) {
	var result IngestObjectRequest
	var hasValues bool
	if cmd.IsSet("tenant-id") {
		hasValues = true
		result.TenantId = cmd.String("tenant-id")
	}
	if cmd.IsSet("table") {
		hasValues = true
		result.Table = cmd.String("table")
	}
	if cmd.IsSet("source") {
		hasValues = true
		result.Source = cmd.String("source")
	}
	if cmd.IsSet("destination") {
		hasValues = true
		result.Destination = cmd.String("destination")
	}
	if cmd.IsSet("location") {
		hasValues = true
		result.Location = cmd.String("location")
	}
	if !hasValues {
		return nil, nil
	}
	return &result, nil
}