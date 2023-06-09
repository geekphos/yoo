package user

import (
	"context"
	"phos.cc/yoo/pkg/auth"
	"phos.cc/yoo/pkg/token"
	"regexp"
	"time"

	"github.com/jinzhu/copier"

	"phos.cc/yoo/internal/pkg/errno"
	"phos.cc/yoo/internal/pkg/model"
	"phos.cc/yoo/internal/yoo/store"
	v1 "phos.cc/yoo/pkg/api/yoo/v1"
)

type UserBiz interface {
	ChangePassword(ctx context.Context, email string, r *v1.ChangePasswordRequest) error
	Create(ctx context.Context, r *v1.CreateUserRequest) error
	Login(ctx context.Context, r *v1.LoginRequest) (*v1.LoginResponse, error)
	Refresh(ctx context.Context, email string) (*v1.RefreshResponse, error)
	Profile(ctx context.Context, email string) (*v1.ProfileResponse, error)
}

type userBiz struct {
	ds store.IStore
}

var _ UserBiz = (*userBiz)(nil)

func New(ds store.IStore) *userBiz {
	return &userBiz{ds: ds}
}

func (b *userBiz) Create(ctx context.Context, r *v1.CreateUserRequest) error {
	var userM = &model.UserM{}
	_ = copier.CopyWithOption(userM, r, copier.Option{IgnoreEmpty: true})

	if err := b.ds.Users().Create(ctx, userM); err != nil {
		if match, _ := regexp.MatchString("Duplicate entry '.*' for key 'email'", err.Error()); match {
			return errno.ErrUserAlreadyExist
		}
		return err
	}

	return nil
}

func (b *userBiz) Login(ctx context.Context, r *v1.LoginRequest) (*v1.LoginResponse, error) {
	user, err := b.ds.Users().GetByEmail(ctx, r.Email)
	if err != nil {
		return nil, errno.ErrUserNotFound
	}

	// compare password
	if err := auth.Compare(user.Password, r.Password); err != nil {
		return nil, errno.ErrPasswordIncorrect
	}

	// generate token
	accessToken, err := token.Sign(user.Email, int(user.ID), token.AccessToken)
	if err != nil {
		return nil, errno.ErrSignToken
	}

	// generate refresh token
	refreshToken, err := token.Sign(user.Email, int(user.ID), token.RefreshToken, token.WithExpDuration(7*24*time.Hour))
	if err != nil {
		return nil, errno.ErrSignToken
	}

	return &v1.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
	}, nil
}

func (b *userBiz) ChangePassword(ctx context.Context, email string, r *v1.ChangePasswordRequest) error {
	userM, err := b.ds.Users().GetByEmail(ctx, email)
	if err != nil {
		return errno.ErrUserNotFound
	}

	if err := auth.Compare(userM.Password, r.OldPassword); err != nil {
		return errno.ErrPasswordIncorrect
	}

	userM.Password, _ = auth.Encrypt(r.NewPassword)
	if err := b.ds.Users().Update(ctx, userM); err != nil {
		return errno.InternalServerError
	}
	return nil
}

func (b *userBiz) Profile(ctx context.Context, email string) (*v1.ProfileResponse, error) {
	userM, err := b.ds.Users().GetByEmail(ctx, email)
	if err != nil {
		return nil, errno.ErrUserNotFound
	}

	var resp = &v1.ProfileResponse{}
	_ = copier.Copy(resp, userM)

	return resp, nil
}

func (b *userBiz) Refresh(ctx context.Context, email string) (*v1.RefreshResponse, error) {
	userM, err := b.ds.Users().GetByEmail(ctx, email)
	if err != nil {
		return nil, errno.ErrUserNotFound
	}

	accessToken, err := token.Sign(userM.Email, int(userM.ID), token.AccessToken)
	if err != nil {
		return nil, errno.ErrSignToken
	}

	refreshToken, err := token.Sign(userM.Email, int(userM.ID), token.RefreshToken, token.WithExpDuration(30*24*time.Hour))
	if err != nil {
		return nil, errno.ErrSignToken
	}

	return &v1.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
	}, nil
}
