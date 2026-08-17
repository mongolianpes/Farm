package users

import (
	"context"
	"errors"
	"mime/multipart"

	"project-farm/internal/images"
	pb "project-farm/internal/users/proto"
)

type userInfo struct {
	ID         int
	name       string
	avatarPath string
}

func RegisterUser(avatarHeader *multipart.FileHeader, login, name, password, interests string) error {
	if err := initService(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	respRegister, err := client.service.Register(ctx, &pb.RegisterRequest{
		Login:     login,
		Name:      name,
		Password:  password,
		Interests: interests,
	})
	if err != nil {
		return err
	}

	if avatarHeader != nil {
		if !images.CheckCurrentFileExtansion(avatarHeader) {
			return errors.New("Попробуйте загрузить картинку в другом формате (png, jpg, webp)")
		}

		avatar, err := avatarHeader.Open()
		if err != nil {
			return err
		}

		pathToImages, err := images.SaveImage(200, 200, avatar)
		if err != nil {
			return errors.New("Попробуйте загрузить другую картинку")
		}

		if _, err := client.service.AddAvatar(ctx, &pb.AddAvatarRequest{
			AvatarPath: pathToImages,
			UserID:     respRegister.UserID,
		}); err != nil {
			return err
		}
	}

	return nil
}

func AuthUser(login, password string) (result userInfo, err error) {
	if err = initService(); err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	var respAuth *pb.AuthResponse
	respAuth, err = client.service.Auth(ctx, &pb.AuthRequest{
		Login:    login,
		Password: password,
	})
	if err != nil {
		return
	}

	result.avatarPath = respAuth.AvatarPath
	result.name = respAuth.UserName
	result.ID = int(respAuth.UserID)

	return
}

func GetUserInfo(login string) (result userInfo, err error) {
	if err = initService(); err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	var resp *pb.GetUserInfoResponse
	resp, err = client.service.GetUserInfo(ctx, &pb.GetUserInfoRequest{
		UserLogin: login,
	})
	if err != nil {
		return
	}

	result.avatarPath = resp.AvatarPath
	result.name = resp.Name
	result.ID = int(resp.UserID)
	return
}
