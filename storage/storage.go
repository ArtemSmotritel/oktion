package storage

import "github.com/artemsmotritel/oktion/types"

type Storage interface {
	GetUserByID(id int64) (*types.User, error)
	GetUsers() ([]types.User, error)
	SaveUser(user *types.User) error
	UpdateUser(id int64, request *types.UserUpdateRequest) error
	DeleteUser(id int64) error
	GetUserByEmail(email string) (*types.User, error)

	GetAuctionsByOwnerId(ownerId int64) ([]types.Auction, error)
	GetOwnerIDByAuctionID(auctionId int64) (int64, error)
	GetAuctionByID(id int64) (*types.Auction, error)
	GetAuctions() ([]types.Auction, error)
	SaveAuction(auction *types.Auction) (*types.Auction, error)
	DeleteAuction(id int64) error

	GetAuctionLotsByAuctionID(auctionId int64) ([]types.AuctionLot, error)
	SaveAuctionLot(auctionLot *types.AuctionLot) (*types.AuctionLot, error)
	GetAuctionLotCount(auctionId int64) (int, error)

	GetCategories() ([]types.Category, error)

	SeedData() error
}
