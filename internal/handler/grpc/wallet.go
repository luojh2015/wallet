package grpc

import (
	"context"

	v1 "github.com/luojh/wallet/api/grpc/v1"
	"github.com/luojh/wallet/internal/service"
	"github.com/luojh/wallet/pkg/errors"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// 确保实现了接口
var _ v1.WalletServiceServer = (*WalletServiceServer)(nil)

// WalletServiceServer 钱包服务实现
type WalletServiceServer struct {
	v1.UnimplementedWalletServiceServer
	walletService *service.WalletService
	authService   service.IAuthService
}

func NewWalletServiceServer(walletService *service.WalletService, authService service.IAuthService) *WalletServiceServer {
	return &WalletServiceServer{
		walletService: walletService,
		authService:   authService,
	}
}

func (s *WalletServiceServer) CreateWallet(ctx context.Context, req *v1.CreateWalletRequest) (*v1.CreateWalletResponse, error) {
	res, err := s.walletService.CreateWallet(ctx, req.Name, req.Passwd)
	if err != nil {
		return nil, err
	}
	return &v1.CreateWalletResponse{WalletId: res.ID}, nil
}

func (s *WalletServiceServer) GetWallet(ctx context.Context, req *v1.GetWalletRequest) (*v1.Wallet, error) {
	res, err := s.walletService.GetWallet(ctx, req.WalletId)
	if err != nil {
		return nil, err
	}
	return &v1.Wallet{Id: res.ID, Name: res.Name, Balance: res.Balance}, nil
}

func (s *WalletServiceServer) DeleteWallet(ctx context.Context, req *v1.DeleteWalletRequest) (*emptypb.Empty, error) {
	return nil, errors.ErrNotImplemented
}

func (s *WalletServiceServer) UpdateWallet(ctx context.Context, req *v1.UpdateWalletRequest) (*emptypb.Empty, error) {
	err := s.walletService.UpdateWallet(ctx, req.WalletId, req.Name, req.Passwd)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *WalletServiceServer) Transfer(ctx context.Context, req *v1.TransferRequest) (*v1.TransferResponse, error) {
	res, err := s.walletService.Transfer(ctx, req.FromWalletId, req.ToWalletId, req.IdempotencyKey, req.Amount)
	if err != nil {
		return nil, err
	}
	return &v1.TransferResponse{TransactionId: res.ID}, nil
}

func (s *WalletServiceServer) GetTransactions(ctx context.Context, req *v1.GetTransactionsRequest) (*v1.GetTransactionsResponse, error) {
	res, total, err := s.walletService.ListTransactions(ctx, req.WalletId, int(req.Offset), int(req.Limit))
	if err != nil {
		return nil, err
	}
	ts := make([]*v1.Transaction, len(res))
	for i, t := range res {
		var tp *timestamppb.Timestamp
		if t.CompletedAt != nil {
			tp = timestamppb.New(*t.CompletedAt)
		}
		ts[i] = &v1.Transaction{
			Id:             t.ID,
			FromWalletId:   t.FromWalletID,
			ToWalletId:     t.ToWalletID,
			Type:           v1.TransactionType(t.Type),
			Amount:         t.Amount,
			Status:         v1.TransactionStatus(t.Status),
			IdempotencyKey: t.IdempotencyKey,
			CompletedAt:    tp,
			CreatedAt:      timestamppb.New(t.CreatedAt),
		}
	}
	return &v1.GetTransactionsResponse{Transactions: ts, Total: int32(total)}, nil
}

func (s *WalletServiceServer) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	res, err := s.authService.Login(ctx, req.WalletId, req.Passwd)
	if err != nil {
		return nil, err
	}
	return &v1.LoginResponse{Token: res.Token, ExpiredAt: timestamppb.New(res.ExpiresAt)}, nil
}

func (s *WalletServiceServer) Logout(ctx context.Context, req *v1.LogoutRequest) (*emptypb.Empty, error) {
	err := s.authService.Logout(ctx, req.Token)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
