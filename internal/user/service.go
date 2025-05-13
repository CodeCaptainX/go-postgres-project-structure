package user

import (
	"fmt"
	types "snack-shop/pkg/model"
	"snack-shop/pkg/responses"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UserCreator interface {
	GetLoginSession(login_session string) (bool, *responses.ErrorResponse)
	Show(userShowRequest UserShowRequest) (*UserResponse, *responses.ErrorResponse)
	ShowOne(user_uuid uuid.UUID) (*UserResponse, *responses.ErrorResponse)
	Create(usreq UserNewRequest) (*UserResponse, *responses.ErrorResponse)
	Update(user_uuid uuid.UUID, usreq UserUpdateRequest) (*UserResponse, *responses.ErrorResponse)
	Delete(user_uuid uuid.UUID) (*UserDeleteResponse, *responses.ErrorResponse)
	GetUserFormCreate() (*UserFormCreateResponse, *responses.ErrorResponse)
	GetUserFormUpdate(user_uuid uuid.UUID) (*UserFormUpdateResponse, *responses.ErrorResponse)
	Update_Password(user_uuid uuid.UUID, usreq UserUpdatePasswordRequest) (*UserUpdatePasswordReponse, *responses.ErrorResponse)
	GetUserBasicInfo() (*UserBasicInfoResponse, *responses.ErrorResponse)
}

type UserService struct {
	userCtx  *types.UserContext
	dbPool   *sqlx.DB
	userRepo UserRepo
}

func NewUserService(u *types.UserContext, db *sqlx.DB) *UserService {
	r := NewUserRepoImpl(u, db)
	return &UserService{
		userCtx:  u,
		dbPool:   db,
		userRepo: r,
	}
}

func (u *UserService) GetLoginSession(login_session string) (bool, *responses.ErrorResponse) {
	fmt.Print("u.userCtx", u.userCtx)
	success, err := u.userRepo.GetLoginSession(login_session)
	if success {
		return success, nil
	} else {
		return false, err
	}
}

func (u *UserService) Show(userShowRequest UserShowRequest) (*UserResponse, *responses.ErrorResponse) {

	success, err := u.userRepo.Show(userShowRequest)
	if err == nil {
		return success, nil
	} else {
		return nil, err
	}
}

func (u *UserService) ShowOne(id uuid.UUID) (*UserResponse, *responses.ErrorResponse) {

	success, err := u.userRepo.ShowOne(id)
	if err == nil {
		return success, nil
	} else {
		return nil, err
	}
}

func (u *UserService) Create(usreq UserNewRequest) (*UserResponse, *responses.ErrorResponse) {

	success, err := u.userRepo.Create(usreq)
	return success, err
}

func (u *UserService) Update(id uuid.UUID, usreq UserUpdateRequest) (*UserResponse, *responses.ErrorResponse) {

	success, err := u.userRepo.Update(id, usreq)
	if err != nil {
		return nil, err
	}
	return success, err
}

func (u *UserService) Delete(user_uuid uuid.UUID) (*UserDeleteResponse, *responses.ErrorResponse) {
	success, err := u.userRepo.Delete(user_uuid)
	return success, err
}

func (u *UserService) GetUserFormCreate() (*UserFormCreateResponse, *responses.ErrorResponse) {

	success, err := u.userRepo.GetUserFormCreate()
	return success, err
}

func (u *UserService) GetUserFormUpdate(user_uuid uuid.UUID) (*UserFormUpdateResponse, *responses.ErrorResponse) {

	success, err := u.userRepo.GetUserFormUpdate(user_uuid)
	return success, err
}

func (u *UserService) Update_Password(user_uuid uuid.UUID, usreq UserUpdatePasswordRequest) (*UserUpdatePasswordReponse, *responses.ErrorResponse) {

	success, err := u.userRepo.Update_Password(user_uuid, usreq)
	if err != nil {
		return nil, err
	}
	return success, err
}

func (u *UserService) GetUserBasicInfo() (*UserBasicInfoResponse, *responses.ErrorResponse) {

	success, err := u.userRepo.GetUserBasicInfo(u.userCtx.UserName)
	if err != nil {
		return nil, err
	}
	return success, nil
}
