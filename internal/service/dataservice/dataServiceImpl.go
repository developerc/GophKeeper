package dataservice

import (
	"context"
	"encoding/json"

	"github.com/developerc/GophKeeper/internal/config"
	"github.com/developerc/GophKeeper/internal/entity"
	"github.com/developerc/GophKeeper/internal/repositories/datarepository"
	"github.com/developerc/GophKeeper/internal/security"
	"go.uber.org/zap"
)

// StorageService интерфейс сервиса хранилища
type StorageService interface {
	SaveRawData(ctx context.Context, name, data, userID string) error
	GetRawData(ctx context.Context, name, userID string) (string, error)

	SaveLoginWithPassword(ctx context.Context, name, login, password, userID string) error
	GetLoginWithPassword(ctx context.Context, name, userID string) (entity.CredentialsDTO, error)

	SaveBinaryData(ctx context.Context, name string, data []byte, userID string) error
	GetBinaryData(ctx context.Context, name, userID string) ([]byte, error)

	SaveCardData(ctx context.Context, name string, cardData entity.CardDataDTO, userID string) error
	GetCardData(ctx context.Context, name, userID string) (entity.CardDataDTO, error)

	GetAllSavedDataNames(ctx context.Context, userID string) ([]string, error)

	DelDataByNameUserId(ctx context.Context, name, userID string) error

	UpdRawData(ctx context.Context, name, data, userID string) error
	UpdLoginWithPassword(ctx context.Context, name, login, password, userID string) error
	UpdBinaryData(ctx context.Context, name string, data []byte, userID string) error
	UpdCardData(ctx context.Context, name string, cardData entity.CardDataDTO, userID string) error
}

// StorageService экземпляр сервиса хранилища
var _ StorageService = &storageServiceImpl{}

// storageServiceImpl структура сервиса хранилища
type storageServiceImpl struct {
	rawDataRepository datarepository.RawDataRepository
	cipherManager     *security.CipherManager
}

// SaveRawData метод для сохранения произвольных текстовых данных
func (s storageServiceImpl) SaveRawData(ctx context.Context, name, data, userID string) error {
	config.ServerSettingsGlob.Logger.Info("SaveRawData", zap.String("dataservice", "save raw data"))
	return s.encryptAndSaveData(ctx, name, userID, []byte(data), entity.RAW)
}

// GetRawData метод для получения произвольных текстовых данных
func (s storageServiceImpl) GetRawData(ctx context.Context, name, userID string) (string, error) {
	config.ServerSettingsGlob.Logger.Info("GetRawData", zap.String("dataservice", "get raw data"))

	decryptData, err := s.getAndDecryptData(ctx, name, userID, entity.RAW)
	if err != nil {
		return "", err
	}

	return string(decryptData), nil
}

// SaveLoginWithPassword метод для сохранения логина и пароля
func (s storageServiceImpl) SaveLoginWithPassword(ctx context.Context, name, login, password, userID string) error {
	config.ServerSettingsGlob.Logger.Info("SaveLoginWithPassword", zap.String("dataservice", "save login with password"))
	cred := entity.CredentialsDTO{
		Login:    login,
		Password: password,
	}

	marshalledCred, err := json.Marshal(cred)
	if err != nil {
		return err
	}
	return s.encryptAndSaveData(ctx, name, userID, marshalledCred, entity.CRED)
}

// GetLoginWithPassword метод для получения логина и пароля
func (s storageServiceImpl) GetLoginWithPassword(ctx context.Context, name, userID string) (entity.CredentialsDTO, error) {
	config.ServerSettingsGlob.Logger.Info("GetLoginWithPassword", zap.String("dataservice", "get credentials"))

	decryptData, err := s.getAndDecryptData(ctx, name, userID, entity.CRED)
	if err != nil {
		return entity.CredentialsDTO{}, err
	}

	cred := entity.CredentialsDTO{}
	if err := json.Unmarshal(decryptData, &cred); err != nil {
		return entity.CredentialsDTO{}, err
	}
	return cred, nil
}

// SaveBinaryData метод для сохранения бинарных данных
func (s storageServiceImpl) SaveBinaryData(ctx context.Context, name string, data []byte, userID string) error {
	config.ServerSettingsGlob.Logger.Info("SaveBinaryData", zap.String("dataservice", "save binary data"))
	return s.encryptAndSaveData(ctx, name, userID, data, entity.FILE)
}

// GetBinaryData метод для получения логина и пароля
func (s storageServiceImpl) GetBinaryData(ctx context.Context, name, userID string) ([]byte, error) {

	config.ServerSettingsGlob.Logger.Info("GetBinaryData", zap.String("dataservice", "get binary data"))
	return s.getAndDecryptData(ctx, name, userID, entity.FILE)
}

// SaveCardData метод для сохранения данных банковской карты
func (s storageServiceImpl) SaveCardData(ctx context.Context, name string, cardData entity.CardDataDTO, userID string) error {
	config.ServerSettingsGlob.Logger.Info("SaveCardData", zap.String("dataservice", "save card data"))

	marshalledCardData, err := json.Marshal(cardData)
	if err != nil {
		return err
	}

	return s.encryptAndSaveData(ctx, name, userID, marshalledCardData, entity.CARD)
}

// GetCardData метод для получения данных банковской карты
func (s storageServiceImpl) GetCardData(ctx context.Context, name, userID string) (entity.CardDataDTO, error) {
	config.ServerSettingsGlob.Logger.Info("GetCardData", zap.String("dataservice", "get card data"))

	decryptData, err := s.getAndDecryptData(ctx, name, userID, entity.CARD)
	if err != nil {
		return entity.CardDataDTO{}, err
	}

	card := entity.CardDataDTO{}
	if err := json.Unmarshal(decryptData, &card); err != nil {
		return entity.CardDataDTO{}, err
	}

	return card, nil
}

// GetAllSavedDataNames метод для получения всех названий сохранений
func (s storageServiceImpl) GetAllSavedDataNames(ctx context.Context, userID string) ([]string, error) {
	config.ServerSettingsGlob.Logger.Info("GetAllSavedDataNames", zap.String("dataservice", "get data names"))
	return s.rawDataRepository.GetAllSavedDataNames(ctx, userID)
}

// DelDataByNameUserId метод для удаления записи по name и userID
func (s storageServiceImpl) DelDataByNameUserId(ctx context.Context, name, userID string) error {
	config.ServerSettingsGlob.Logger.Info("DelDataByNameUserId", zap.String("storageServiceImpl", "delete data from db"))
	return s.rawDataRepository.DelDataByNameUserId(ctx, name, userID)
}

// SaveRawData метод для обновления произвольных текстовых данных
func (s storageServiceImpl) UpdRawData(ctx context.Context, name, data, userID string) error {
	config.ServerSettingsGlob.Logger.Info("UpdRawData", zap.String("dataservice", "update raw data"))
	//return s.encryptAndSaveData(ctx, name, userID, []byte(data), entity.RAW)
	return s.encryptAndUpdateData(ctx, name, userID, []byte(data), entity.RAW)
}

// UpdLoginWithPassword метод для обновления логина и пароля
func (s storageServiceImpl) UpdLoginWithPassword(ctx context.Context, name, login, password, userID string) error {
	config.ServerSettingsGlob.Logger.Info("UpdLoginWithPassword", zap.String("dataservice", "update login with password"))
	cred := entity.CredentialsDTO{
		Login:    login,
		Password: password,
	}

	marshalledCred, err := json.Marshal(cred)
	if err != nil {
		return err
	}
	return s.encryptAndUpdateData(ctx, name, userID, marshalledCred, entity.CRED)
}

// UpdBinaryData метод для обновления бинарных данных
func (s storageServiceImpl) UpdBinaryData(ctx context.Context, name string, data []byte, userID string) error {
	config.ServerSettingsGlob.Logger.Info("UpdBinaryData", zap.String("dataservice", "update binary data"))
	return s.encryptAndUpdateData(ctx, name, userID, data, entity.FILE)
}

// UpdCardData метод для обновления данных банковской карты
func (s storageServiceImpl) UpdCardData(ctx context.Context, name string, cardData entity.CardDataDTO, userID string) error {
	config.ServerSettingsGlob.Logger.Info("UpdCardData", zap.String("dataservice", "update card data"))

	marshalledCardData, err := json.Marshal(cardData)
	if err != nil {
		return err
	}

	return s.encryptAndUpdateData(ctx, name, userID, marshalledCardData, entity.CARD)
}

func (s storageServiceImpl) encryptAndSaveData(
	ctx context.Context,
	name, userID string,
	data []byte,
	dataType entity.DataType) error {

	savedData, err := s.cipherManager.Encrypt(data)
	if err != nil {
		return err
	}
	return s.rawDataRepository.Save(ctx, userID, name, savedData, dataType)
}

func (s storageServiceImpl) encryptAndUpdateData(
	ctx context.Context,
	name, userID string,
	data []byte,
	dataType entity.DataType) error {

	savedData, err := s.cipherManager.Encrypt(data)
	if err != nil {
		return err
	}
	//return s.rawDataRepository.Save(ctx, userID, name, savedData, dataType)
	return s.rawDataRepository.Update(ctx, userID, name, savedData, dataType)
}

func (s storageServiceImpl) getAndDecryptData(
	ctx context.Context,
	name, userID string,
	dataType entity.DataType) ([]byte, error) {

	data, err := s.rawDataRepository.GetByNameAndTypeAndUserID(ctx, userID, name, dataType)
	if err != nil {
		return nil, err
	}

	decryptData, err := s.cipherManager.Decrypt(data)
	if err != nil {
		return nil, err
	}

	return decryptData, nil
}

// New конструктор UserService
func New(rawDataRepository datarepository.RawDataRepository, cipherManager *security.CipherManager) StorageService {
	return &storageServiceImpl{
		rawDataRepository,
		cipherManager,
	}
}
