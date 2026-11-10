package testSupport

import (
	channelrepo "github.com/wichat/wichat/backend/ms-user/internal/channel/repository"
	channelmock "github.com/wichat/wichat/backend/ms-user/internal/channel/repository/mock"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
)

// ChannelRepo returns the in-memory channel repository sharing test state.
func ChannelRepo(st *usermock.State) channelrepo.PgRepository {
	return channelmock.NewPgRepositoryMock(st)
}
