package arangodb

import (
	"context"
	models "github.com/Nubes3/nubes3-user-service/pkg/models"
	"github.com/Nubes3/nubes3-user-service/pkg/utils"
	"github.com/arangodb/go-driver"
	scrypt "github.com/elithrar/simple-scrypt"
	"github.com/thanhpk/randstr"
	"strings"
	"time"
)

func CreateUser(
	firstname string,
	lastname string,
	username string,
	password string,
	email string,
	dob time.Time,
	company string,
	gender bool,
) (*models.User, error) {
	createdTime := time.Now()
	passwordHashed, err := scrypt.GenerateFromPassword([]byte(password), scrypt.DefaultParams)
	if err != nil {
		return nil, err
	}

	otp := utils.GenerateOtp()

	doc := models.User{
		Firstname:    firstname,
		Lastname:     lastname,
		Username:     username,
		Pass:         string(passwordHashed),
		Email:        email,
		Dob:          dob,
		Company:      company,
		Gender:       gender,
		IsActive:     false,
		RefreshToken: strings.ToUpper(randstr.Hex(8)),
		Otp:          otp,
		IsBanned:     false,
		CreatedAt:    createdTime,
		UpdatedAt:    createdTime,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*ContextExpiredTime)
	defer cancel()

	usernameDup, emailDup, err := isUserDuplicate(username, email)
	if err != nil {
		return nil, &utils.ModelError{
			Msg:     err.Error(),
			ErrType: utils.DbError,
		}
	}

	if usernameDup {
		return nil, &utils.ModelError{
			Msg:     "username duplicate",
			ErrType: utils.Duplicated,
		}
	}

	if emailDup {
		return nil, &utils.ModelError{
			Msg:     "email duplicate",
			ErrType: utils.Duplicated,
		}
	}

	meta, err := userCol.CreateDocument(ctx, doc)
	if err != nil {
		return nil, &utils.ModelError{
			Msg:     err.Error(),
			ErrType: utils.DbError,
		}
	}

	//LOG CREATE USER
	//_ = nats.SendUserEvent(doc.Firstname, doc.Lastname, doc.Username,
	//	doc.Pass, doc.Email, doc.Dob, doc.Company, doc.Gender, doc.IsActive, doc.IsBanned,
	//	"create")

	return &models.User{
		Id:        meta.Key,
		Firstname: doc.Firstname,
		Lastname:  doc.Lastname,
		Username:  doc.Username,
		Pass:      doc.Pass,
		Email:     doc.Email,
		Dob:       doc.Dob,
		Company:   doc.Company,
		Gender:    doc.Gender,
		IsActive:  doc.IsActive,
		IsBanned:  doc.IsBanned,
		CreatedAt: doc.CreatedAt,
		UpdatedAt: doc.UpdatedAt,
	}, nil
}

func FindUserById(uid string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*ContextExpiredTime)
	defer cancel()

	user := models.User{}
	meta, err := userCol.ReadDocument(ctx, uid, &user)
	if err != nil {
		if driver.IsNotFound(err) {
			return nil, &utils.ModelError{
				Msg:     "user not found",
				ErrType: utils.NotFound,
			}
		}

		return nil, &utils.ModelError{
			Msg:     err.Error(),
			ErrType: utils.DbError,
		}
	}

	user.Id = meta.Key

	return &user, nil
}

func FindUserByUsername(uname string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*ContextExpiredTime)
	defer cancel()

	query := "FOR u IN users FILTER u.username == @uname LIMIT 1 RETURN u"
	bindVars := map[string]interface{}{
		"uname": uname,
	}

	user := models.User{}
	cursor, err := arangoDb.Query(ctx, query, bindVars)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()

	for {
		meta, err := cursor.ReadDocument(ctx, &user)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, err
		}
		user.Id = meta.Key
	}

	if user.Id == "" {
		return nil, &utils.ModelError{
			Msg:     "user not found",
			ErrType: utils.NotFound,
		}
	}

	return &user, nil
}

func FindUserByEmail(mail string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*ContextExpiredTime)
	defer cancel()

	query := "FOR u IN users FILTER u.email == @email LIMIT 1 RETURN u"
	bindVars := map[string]interface{}{
		"email": mail,
	}

	user := models.User{}
	cursor, err := arangoDb.Query(ctx, query, bindVars)
	if err != nil {
		return nil, &utils.ModelError{
			Msg:     err.Error(),
			ErrType: utils.DbError,
		}
	}
	defer cursor.Close()

	for {
		meta, err := cursor.ReadDocument(ctx, &user)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, &utils.ModelError{
				Msg:     err.Error(),
				ErrType: utils.DbError,
			}
		}
		user.Id = meta.Key
	}

	if user.Id == "" {
		return nil, &utils.ModelError{
			Msg:     "user not found",
			ErrType: utils.NotFound,
		}
	}

	return &user, nil
}

func UpdateUserData(
	uid string,
	firstname string,
	lastname string,
	dob time.Time,
	company string,
	gender bool) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*ContextExpiredTime)
	defer cancel()

	updatedTime := time.Now()

	query := "FOR u IN users FILTER u._key == @uid UPDATE u " +
		"WITH { firstname: @firstname, " +
		"lastname: @lastname, " +
		"dob: @dob, " +
		"company: @company, " +
		"gender: @gender, " +
		"updated_at: @updatedAt } " +
		"IN users RETURN NEW"
	bindVars := map[string]interface{}{
		"uid":       uid,
		"firstname": firstname,
		"lastname":  lastname,
		"dob":       dob,
		"company":   company,
		"gender":    gender,
		"updatedAt": updatedTime,
	}

	cursor, err := arangoDb.Query(ctx, query, bindVars)
	if err != nil {
		return nil, &utils.ModelError{
			Msg:     err.Error(),
			ErrType: utils.DbError,
		}
	}
	defer cursor.Close()

	user := models.User{}
	for {
		meta, err := cursor.ReadDocument(ctx, &user)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, &utils.ModelError{
				Msg:     err.Error(),
				ErrType: utils.DbError,
			}
		}
		user.Id = meta.Key
	}

	if user.Id == "" {
		return nil, &utils.ModelError{
			Msg:     "user not found",
			ErrType: utils.NotFound,
		}
	}

	//LOG UPDATE USER
	//_ = nats.SendUserEvent(user.Firstname, user.Lastname, user.Username,
	//	user.Pass, user.Email, user.Dob, user.Company, user.Gender, user.IsActive, user.IsBanned,
	//	"update")

	return &user, err
}

func UpdateActive(uname string, isActive bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*ContextExpiredTime)
	defer cancel()

	query := "FOR u IN users FILTER u.username == @uname " +
		"UPDATE u WITH { is_active: @isActive } IN users RETURN NEW"
	bindVars := map[string]interface{}{
		"uname":    uname,
		"isActive": isActive,
	}

	user := models.User{}
	cursor, err := arangoDb.Query(ctx, query, bindVars)
	if err != nil {
		return err
	}
	defer cursor.Close()

	for {
		_, err := cursor.ReadDocument(ctx, &user)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return err
		}
	}

	//user, err := FindUserByUsername(uname)
	//if err != nil {
	//	return &models.ModelError{
	//		Msg:     "user not found",
	//		ErrType: models.DocumentNotFound,
	//	}
	//}
	//userUpdate := User{
	//	IsActive: isActive,
	//}
	//
	//meta, err := userCol.UpdateDocument(ctx, user.Id, &userUpdate)
	//
	//if err != nil {
	//	if driver.IsNotFound(err) {
	//		return &models.ModelError{
	//			Msg:     "user not found",
	//			ErrType: models.DocumentNotFound,
	//		}
	//	}
	//
	//	return &models.ModelError{
	//		Msg:     err.Error(),
	//		ErrType: models.DbError,
	//	}
	//}
	//user.Id = meta.Key
	return err
}

func UpdateUserPassword(uid string, password string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*ContextExpiredTime)
	defer cancel()

	passwordHashed, err := scrypt.GenerateFromPassword([]byte(password), scrypt.DefaultParams)
	if err != nil {
		return nil, err
	}
	user := models.User{
		Pass: string(passwordHashed),
	}

	meta, err := userCol.UpdateDocument(ctx, uid, &user)
	if err != nil {
		if driver.IsNotFound(err) {
			return nil, &utils.ModelError{
				Msg:     "user not found",
				ErrType: utils.NotFound,
			}
		}

		return nil, &utils.ModelError{
			Msg:     err.Error(),
			ErrType: utils.DbError,
		}
	}

	user.Id = meta.Key
	return &user, err
}

func UpdateOtp(username string) (*models.User, error) {
	otp := utils.GenerateOtp()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*ContextExpiredTime)
	defer cancel()

	query := "FOR u IN users FILTER u.username == @username LIMIT 1 " +
		"UPDATE u WITH { otp: @otp } IN users RETURN NEW"
	bindVars := map[string]interface{}{
		"otp":      otp,
		"username": username,
	}

	user := models.User{}
	cursor, err := arangoDb.Query(ctx, query, bindVars)
	if err != nil {
		return nil, &utils.ModelError{
			Msg:     "not found",
			ErrType: utils.NotFound,
		}
	}
	defer cursor.Close()

	for {
		meta, err := cursor.ReadDocument(ctx, &user)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, &utils.ModelError{
				Msg:     "not found",
				ErrType: utils.NotFound,
			}
		}
		user.Id = meta.Key
	}

	return &user, nil
}

func ConfirmOtp(username, otp string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*ContextExpiredTime)
	defer cancel()

	query := "FOR u IN users FILTER u.username == @username AND u.otp.otp == @otpVal AND u.otp.expired < @deadline LIMIT 1 " +
		"UPDATE u WITH { otp: null } IN users RETURN NEW"
	bindVars := map[string]interface{}{
		"username": username,
		"otpVal":   otp,
		"deadline": time.Now(),
	}

	user := models.User{}
	cursor, err := arangoDb.Query(ctx, query, bindVars)
	if err != nil {
		return nil, &utils.ModelError{
			Msg:     err.Error(),
			ErrType: utils.DbError,
		}
	}
	defer cursor.Close()

	for {
		meta, err := cursor.ReadDocument(ctx, &user)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, &utils.ModelError{
				Msg:     err.Error(),
				ErrType: utils.DbError,
			}
		}
		user.Id = meta.Key
	}

	if user.Id == "" {
		return nil, &utils.ModelError{
			Msg:     "otp expired or mismatched",
			ErrType: utils.NotFound,
		}
	}

	return &user, nil
}

func UpdateRefreshToken(uid string) (*models.User, error) {
	newRefreshToken := strings.ToUpper(randstr.Hex(8))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*ContextExpiredTime)
	defer cancel()

	query := "FOR u IN users FILTER u._key == @uid LIMIT 1 " +
		"UPDATE u WITH { refresh_token: @rf } IN users RETURN NEW"
	bindVars := map[string]interface{}{
		"uid": uid,
		"rf":  newRefreshToken,
	}

	user := models.User{}
	cursor, err := arangoDb.Query(ctx, query, bindVars)
	if err != nil {
		return nil, &utils.ModelError{
			Msg:     "not found",
			ErrType: utils.NotFound,
		}
	}
	defer cursor.Close()

	for {
		meta, err := cursor.ReadDocument(ctx, &user)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, &utils.ModelError{
				Msg:     "not found",
				ErrType: utils.NotFound,
			}
		}
		user.Id = meta.Key
	}

	return &user, nil
}

func UpdateBanStatus(uid string, isBan bool) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*ContextExpiredTime)
	defer cancel()

	user := models.User{
		IsBanned: isBan,
	}

	meta, err := userCol.UpdateDocument(ctx, uid, &user)
	if err != nil {
		if driver.IsNotFound(err) {
			return nil, &utils.ModelError{
				Msg:     "user not found",
				ErrType: utils.NotFound,
			}
		}

		return nil, &utils.ModelError{
			Msg:     err.Error(),
			ErrType: utils.DbError,
		}
	}

	user.Id = meta.Key
	return &user, err
}

func RemoveUser(uid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*ContextExpiredTime)
	defer cancel()

	_, err := userCol.RemoveDocument(ctx, uid)
	if err != nil {
		if driver.IsNotFound(err) {
			return &utils.ModelError{
				Msg:     "user not found",
				ErrType: utils.NotFound,
			}
		}

		return &utils.ModelError{
			Msg:     err.Error(),
			ErrType: utils.DbError,
		}
	}

	return nil
}

func isUserDuplicate(username, email string) (usernameDup bool, emailDup bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*ContextExpiredTime)
	defer cancel()

	query := "FOR u IN users FILTER u.username == @username OR u.email == @email LIMIT 1 RETURN u"
	bindVars := map[string]interface{}{
		"email":    email,
		"username": username,
	}

	user := models.User{}
	cursor, err := arangoDb.Query(ctx, query, bindVars)
	if err != nil {
		usernameDup = false
		emailDup = false
		err = &utils.ModelError{
			Msg:     err.Error(),
			ErrType: utils.DbError,
		}
		return
	}
	defer cursor.Close()

	for {
		meta, err := cursor.ReadDocument(ctx, &user)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			usernameDup = false
			emailDup = false
			err = &utils.ModelError{
				Msg:     err.Error(),
				ErrType: utils.DbError,
			}
			return
		}
		user.Id = meta.Key
	}

	if user.Id != "" {
		if user.Email == email {
			emailDup = true
		}
		if user.Username == username {
			usernameDup = true
		}
		err = nil
		return
	}

	return false, false, nil
}
