package service

import (
	"errors"

	"shoe-store/internal/model"
	"shoe-store/internal/repository"
)

type OrderService struct {
	Repo *repository.OrderRepo
}

func NewOrderService(repo *repository.OrderRepo) *OrderService {
	return &OrderService{Repo: repo}
}

func (s *OrderService) List() ([]model.Order, error) {
	return s.Repo.List()
}

func (s *OrderService) GetByID(id int64) (*model.Order, error) {
	o, err := s.Repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, errors.New("order not found")
	}
	return o, nil
}

func validateOrderInput(input model.OrderInput) error {
	if input.StatusID <= 0 {
		return errors.New("statusId is required")
	}
	if input.PickupPointID <= 0 {
		return errors.New("pickupPointId is required")
	}
	if len(input.Items) == 0 {
		return errors.New("items must not be empty")
	}
	return nil
}

func (s *OrderService) Create(input model.OrderInput) (int64, error) {
	if err := validateOrderInput(input); err != nil {
		return 0, err
	}
	return s.Repo.Create(input)
}

func (s *OrderService) Update(id int64, input model.OrderInput) error {
	if err := validateOrderInput(input); err != nil {
		return err
	}
	return s.Repo.Update(id, input)
}

func (s *OrderService) Delete(id int64) error {
	return s.Repo.Delete(id)
}
