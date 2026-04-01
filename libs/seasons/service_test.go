package seasons_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/tobyrushton/padel-stats/libs/fakes"
	"github.com/tobyrushton/padel-stats/libs/seasons"
)

type ServiceTestSuite struct {
	suite.Suite

	repo    *fakes.FakeSeasonsRepository
	service *seasons.Service
	ctx     context.Context
}

func (suite *ServiceTestSuite) SetupTest() {
	suite.repo = new(fakes.FakeSeasonsRepository)
	suite.ctx = context.Background()

	var err error
	suite.service, err = seasons.NewService(suite.repo)
	suite.Require().NoError(err)
}

func (suite *ServiceTestSuite) TestNewService_NilRepository() {
	service, err := seasons.NewService(nil)

	suite.Error(err)
	suite.Nil(service)
	suite.Equal("seasons repository is required", err.Error())
}

func (suite *ServiceTestSuite) TestGetSeasons_Success() {
	expectedSeasons := []*seasons.Season{
		{ID: 1, Name: "Season 1"},
		{ID: 2, Name: "Season 2"},
	}

	suite.repo.GetSeasonsStub = func(ctx context.Context) ([]*seasons.Season, error) {
		return expectedSeasons, nil
	}

	seasons, err := suite.service.GetSeasons(suite.ctx)

	suite.NoError(err)
	suite.Equal(expectedSeasons, seasons)
}

func (suite *ServiceTestSuite) TestGetActiveSeason_Success() {
	expectedSeason := &seasons.Season{ID: 1, Name: "Active Season"}

	suite.repo.GetActiveSeasonStub = func(ctx context.Context) (*seasons.Season, error) {
		return expectedSeason, nil
	}

	activeSeason, err := suite.service.GetActiveSeason(suite.ctx)

	suite.NoError(err)
	suite.Equal(expectedSeason, activeSeason)
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
