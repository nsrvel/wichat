// Package channelcategory is the gRPC client for ms-user category-domain RPCs on ChannelCategoryService.
package channelcategory

import (
	"context"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

// Client calls category-domain ChannelCategoryService RPCs with shared timeouts, retries, and circuit breaking.
type Client interface {
	CreateChannelCategory(ctx context.Context, in *userv1.CreateChannelCategoryRequest) (*userv1.CreateChannelCategoryResponse, error)
	UpdateChannelCategory(ctx context.Context, in *userv1.UpdateChannelCategoryRequest) (*userv1.UpdateChannelCategoryResponse, error)
	DeleteChannelCategory(ctx context.Context, in *userv1.DeleteChannelCategoryRequest) (*userv1.DeleteChannelCategoryResponse, error)
	ListChannelCategoryPermissions(ctx context.Context, in *userv1.ListChannelCategoryPermissionsRequest) (*userv1.ListChannelCategoryPermissionsResponse, error)
	UpsertChannelCategoryPermission(ctx context.Context, in *userv1.UpsertChannelCategoryPermissionRequest) (*userv1.UpsertChannelCategoryPermissionResponse, error)
	DeleteChannelCategoryPermission(ctx context.Context, in *userv1.DeleteChannelCategoryPermissionRequest) (*userv1.DeleteChannelCategoryPermissionResponse, error)
	ReorderChannelCategories(ctx context.Context, in *userv1.ReorderChannelCategoriesRequest) (*userv1.ReorderChannelCategoriesResponse, error)
}

type client struct {
	stub  userv1.ChannelCategoryServiceClient
	log   logger.Logger
	read  utils.CallConfig
	write utils.CallConfig
}

// New wraps an existing ChannelCategoryService client stub.
func New(stub userv1.ChannelCategoryServiceClient, settings utils.Settings, log logger.Logger) Client {
	breaker := utils.NewBreaker(settings)
	read := utils.CallConfigFromSettings(settings, utils.RetryIdempotent)
	read.Breaker = breaker
	write := utils.CallConfigFromSettings(settings, utils.RetryNone)
	write.Breaker = breaker

	return &client{
		stub:  stub,
		log:   log,
		read:  read,
		write: write,
	}
}

func (c *client) CreateChannelCategory(ctx context.Context, in *userv1.CreateChannelCategoryRequest) (*userv1.CreateChannelCategoryResponse, error) {
	return utils.CallUnaryResult(ctx, c.write, c.log, "channelcategory.ChannelCategoryService/CreateChannelCategory", func(callCtx context.Context) (*userv1.CreateChannelCategoryResponse, error) {
		return c.stub.CreateChannelCategory(callCtx, in)
	})
}

func (c *client) UpdateChannelCategory(ctx context.Context, in *userv1.UpdateChannelCategoryRequest) (*userv1.UpdateChannelCategoryResponse, error) {
	return utils.CallUnaryResult(ctx, c.write, c.log, "channelcategory.ChannelCategoryService/UpdateChannelCategory", func(callCtx context.Context) (*userv1.UpdateChannelCategoryResponse, error) {
		return c.stub.UpdateChannelCategory(callCtx, in)
	})
}

func (c *client) DeleteChannelCategory(ctx context.Context, in *userv1.DeleteChannelCategoryRequest) (*userv1.DeleteChannelCategoryResponse, error) {
	return utils.CallUnaryResult(ctx, c.write, c.log, "channelcategory.ChannelCategoryService/DeleteChannelCategory", func(callCtx context.Context) (*userv1.DeleteChannelCategoryResponse, error) {
		return c.stub.DeleteChannelCategory(callCtx, in)
	})
}

func (c *client) ListChannelCategoryPermissions(ctx context.Context, in *userv1.ListChannelCategoryPermissionsRequest) (*userv1.ListChannelCategoryPermissionsResponse, error) {
	return utils.CallUnaryResult(ctx, c.read, c.log, "channelcategory.ChannelCategoryService/ListChannelCategoryPermissions", func(callCtx context.Context) (*userv1.ListChannelCategoryPermissionsResponse, error) {
		return c.stub.ListChannelCategoryPermissions(callCtx, in)
	})
}

func (c *client) UpsertChannelCategoryPermission(ctx context.Context, in *userv1.UpsertChannelCategoryPermissionRequest) (*userv1.UpsertChannelCategoryPermissionResponse, error) {
	return utils.CallUnaryResult(ctx, c.write, c.log, "channelcategory.ChannelCategoryService/UpsertChannelCategoryPermission", func(callCtx context.Context) (*userv1.UpsertChannelCategoryPermissionResponse, error) {
		return c.stub.UpsertChannelCategoryPermission(callCtx, in)
	})
}

func (c *client) DeleteChannelCategoryPermission(ctx context.Context, in *userv1.DeleteChannelCategoryPermissionRequest) (*userv1.DeleteChannelCategoryPermissionResponse, error) {
	return utils.CallUnaryResult(ctx, c.write, c.log, "channelcategory.ChannelCategoryService/DeleteChannelCategoryPermission", func(callCtx context.Context) (*userv1.DeleteChannelCategoryPermissionResponse, error) {
		return c.stub.DeleteChannelCategoryPermission(callCtx, in)
	})
}

func (c *client) ReorderChannelCategories(ctx context.Context, in *userv1.ReorderChannelCategoriesRequest) (*userv1.ReorderChannelCategoriesResponse, error) {
	return utils.CallUnaryResult(ctx, c.write, c.log, "channelcategory.ChannelCategoryService/ReorderChannelCategories", func(callCtx context.Context) (*userv1.ReorderChannelCategoriesResponse, error) {
		return c.stub.ReorderChannelCategories(callCtx, in)
	})
}
