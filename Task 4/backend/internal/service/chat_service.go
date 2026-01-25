package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"webchat/internal/models"
	"webchat/internal/repository"
)

type ChatService interface {
	CreateChat(req *models.CreateChatRequest, creatorID string) (*models.Chat, error)
	GetUserChats(userID string) ([]models.Chat, error)
	GetPublicChats() ([]models.Chat, error)
	GetChat(chatID, userID string) (*models.Chat, error)
	UpdateChat(chatID string, req *models.UpdateChatRequest, userID string) error
	CreateInvite(chatID, userID string, req *models.CreateInviteRequest) (*models.Invitation, error)
	JoinChat(inviteCode, userID string) error
	AssignRole(chatID, assignerID string, req *models.AssignRoleRequest) error
	RemoveMember(chatID, removerID, memberID string) error
	GetChatMembers(chatID, userID string) ([]models.ChatMember, error)
}

type chatService struct {
	chatRepo repository.ChatRepository
	userRepo repository.UserRepository
}

func NewChatService(chatRepo repository.ChatRepository, userRepo repository.UserRepository) ChatService {
	return &chatService{
		chatRepo: chatRepo,
		userRepo: userRepo,
	}
}

func (s *chatService) CreateChat(req *models.CreateChatRequest, creatorID string) (*models.Chat, error) {
	chat := &models.Chat{
		Name:            req.Name,
		Description:     req.Description,
		IsPublic:        req.IsPublic,
		CreatorID:       creatorID,
		OnlyAdminInvite: req.OnlyAdminInvite,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	err := s.chatRepo.CreateChat(chat)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания чата: %w", err)
	}

	return chat, nil
}

func (s *chatService) GetUserChats(userID string) ([]models.Chat, error) {
	chats, err := s.chatRepo.GetUserChats(userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения чатов: %w", err)
	}

	return chats, nil
}

func (s *chatService) GetPublicChats() ([]models.Chat, error) {
	chats, err := s.chatRepo.GetPublicChats()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения публичных чатов: %w", err)
	}

	return chats, nil
}

func (s *chatService) GetChat(chatID, userID string) (*models.Chat, error) {
	if !s.isPublicChat(chatID) {
		inChat, err := s.chatRepo.CheckUserInChat(chatID, userID)
		if err != nil {
			return nil, fmt.Errorf("ошибка проверки доступа: %w", err)
		}
		if !inChat {
			return nil, errors.New("доступ запрещен")
		}
	}

	chat, err := s.chatRepo.GetChatByID(chatID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения чата: %w", err)
	}

	return chat, nil
}

func (s *chatService) UpdateChat(chatID string, req *models.UpdateChatRequest, userID string) error {
	hasPermission, err := s.chatRepo.UserHasPermission(chatID, userID, "can_remove_users")
	if err != nil {
		return fmt.Errorf("ошибка проверки прав: %w", err)
	}
	if !hasPermission {
		return errors.New("недостаточно прав")
	}

	err = s.chatRepo.UpdateChat(chatID, req)
	if err != nil {
		return fmt.Errorf("ошибка обновления чата: %w", err)
	}

	return nil
}

func (s *chatService) CreateInvite(chatID, userID string, req *models.CreateInviteRequest) (*models.Invitation, error) {
	chat, err := s.chatRepo.GetChatByID(chatID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения чата: %w", err)
	}

	if chat.OnlyAdminInvite {
		hasPermission, err := s.chatRepo.UserHasPermission(chatID, userID, "can_remove_users")
		if err != nil {
			return nil, fmt.Errorf("ошибка проверки прав: %w", err)
		}
		if !hasPermission {
			return nil, errors.New("только администраторы могут приглашать")
		}
	}

	inviteCode := strings.ReplaceAll(uuid.New().String(), "-", "")[:10]

	invite := &models.Invitation{
		ChatID:     chatID,
		InviteCode: inviteCode,
		IsUsed:     false,
		ExpiresAt:  time.Now().Add(time.Duration(req.ExpiresInHours) * time.Hour),
		CreatedAt:  time.Now(),
	}

	err = s.chatRepo.CreateInvitation(invite)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания приглашения: %w", err)
	}

	return invite, nil
}

func (s *chatService) JoinChat(inviteCode, userID string) error {
	invite, err := s.chatRepo.GetInvitationByCode(inviteCode)
	if err != nil {
		return fmt.Errorf("ошибка проверки приглашения: %w", err)
	}

	if invite == nil {
		return errors.New("приглашение не найдено или истекло")
	}

	if invite.IsUsed {
		return errors.New("приглашение уже использовано")
	}

	err = s.chatRepo.UseInvitation(inviteCode, userID)
	if err != nil {
		return fmt.Errorf("ошибка вступления в чат: %w", err)
	}

	return nil
}

func (s *chatService) AssignRole(chatID, assignerID string, req *models.AssignRoleRequest) error {
	canAssign, err := s.chatRepo.UserHasPermission(chatID, assignerID, "can_assign_roles")
	if err != nil {
		return fmt.Errorf("ошибка проверки прав: %w", err)
	}
	if !canAssign {
		return errors.New("недостаточно прав для назначения ролей")
	}

	inChat, err := s.chatRepo.CheckUserInChat(chatID, req.UserID)
	if err != nil {
		return fmt.Errorf("ошибка проверки участника: %w", err)
	}
	if !inChat {
		return errors.New("пользователь не участник чата")
	}

	role := &models.ChatRole{
		UserID:            req.UserID,
		RoleName:          req.RoleName,
		CanDeleteMessages: req.CanDeleteMessages,
		CanRemoveUsers:    req.CanRemoveUsers,
		CanAssignRoles:    req.CanAssignRoles,
	}

	err = s.chatRepo.SetChatRole(chatID, req.UserID, role)
	if err != nil {
		return fmt.Errorf("ошибка назначения роли: %w", err)
	}

	return nil
}

func (s *chatService) RemoveMember(chatID, removerID, memberID string) error {
	if removerID == memberID {
		return errors.New("нельзя удалить самого себя")
	}

	canRemove, err := s.chatRepo.UserHasPermission(chatID, removerID, "can_remove_users")
	if err != nil {
		return fmt.Errorf("ошибка проверки прав: %w", err)
	}
	if !canRemove {
		return errors.New("недостаточно прав для удаления пользователей")
	}

	memberRole, err := s.chatRepo.GetChatRole(chatID, memberID)
	if err != nil {
		return fmt.Errorf("ошибка проверки роли: %w", err)
	}

	removerRole, err := s.chatRepo.GetChatRole(chatID, removerID)
	if err != nil {
		return fmt.Errorf("ошибка проверки роли: %w", err)
	}

	if memberRole.RoleName == "admin" && removerRole.RoleName != "owner" {
		return errors.New("только владелец может удалять администраторов")
	}

	err = s.chatRepo.RemoveChatMember(chatID, memberID)
	if err != nil {
		return fmt.Errorf("ошибка удаления пользователя: %w", err)
	}

	return nil
}

func (s *chatService) GetChatMembers(chatID, userID string) ([]models.ChatMember, error) {
	if !s.isPublicChat(chatID) {
		inChat, err := s.chatRepo.CheckUserInChat(chatID, userID)
		if err != nil {
			return nil, fmt.Errorf("ошибка проверки доступа: %w", err)
		}
		if !inChat {
			return nil, errors.New("доступ запрещен")
		}
	}

	members, err := s.chatRepo.GetChatMembers(chatID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения участников: %w", err)
	}

	return members, nil
}

func (s *chatService) isPublicChat(chatID string) bool {
	chat, err := s.chatRepo.GetChatByID(chatID)
	if err != nil || chat == nil {
		return false
	}
	return chat.IsPublic
}
