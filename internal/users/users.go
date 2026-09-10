package users

import (
	"context"
	"errors"
	"log/slog"
	"mime/multipart"

	"project-farm/internal/images"
	pb "project-farm/internal/users/proto"
)

type userInfo struct {
	ID         int
	Name       string
	AvatarPath string
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
			if _, err := client.service.DeleteUser(ctx, &pb.DeleteUserRequest{
				UserID: respRegister.UserID,
			}); err != nil {
				slog.Error("Пользователь не был удален после попытки загрузки картинки в не поддерживаемом формате")
			}
			return errors.New("Попробуйте загрузить картинку в другом формате (png, jpg, webp)")
		}

		avatar, err := avatarHeader.Open()
		if err != nil {
			if _, err := client.service.DeleteUser(ctx, &pb.DeleteUserRequest{
				UserID: respRegister.UserID,
			}); err != nil {
				slog.Error("Пользователь не был удален после не возможности получить картинку")
			}
			return err
		}

		pathToImages, err := images.SaveImage(200, 200, avatar)
		if err != nil {
			if _, err := client.service.DeleteUser(ctx, &pb.DeleteUserRequest{
				UserID: respRegister.UserID,
			}); err != nil {
				slog.Error("Пользователь не был удален после ошибки от микросервиса Images", "error", err)
			}
			return errors.New("Попробуйте загрузить другую картинку")
		}

		if _, err := client.service.AddAvatar(ctx, &pb.AddAvatarRequest{
			AvatarPath: pathToImages,
			UserID:     respRegister.UserID,
		}); err != nil {
			if _, err := client.service.DeleteUser(ctx, &pb.DeleteUserRequest{
				UserID: respRegister.UserID,
			}); err != nil {
				slog.Error("Пользователь не был удален после ошибки от микросервиса Announcements", "error", err)
			}
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

	result.AvatarPath = respAuth.AvatarPath
	result.Name = respAuth.UserName
	result.ID = int(respAuth.UserID)

	return
}

func GetUserInfo(id int, login string) (result userInfo, err error) {
	if err = initService(); err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	var resp *pb.GetUserInfoResponse
	resp, err = client.service.GetUserInfo(ctx, &pb.GetUserInfoRequest{
		UserLogin: login,
		UserID:    int64(id),
	})
	if err != nil {
		slog.Warn("Не удалось получить данный профиля пользователя", "login", login, "error", err)
		return
	}

	result.AvatarPath = resp.AvatarPath
	result.Name = resp.Name
	result.ID = int(resp.UserID)
	return
}

func GetUserID(login string) (int, error) {
	if err := initService(); err != nil {
		return 0, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	resp, err := client.service.GetUserID(ctx, &pb.GetUserIDRequest{
		Login: login,
	})
	if err != nil {
		return 0, err
	}

	return int(resp.ID), nil
}
