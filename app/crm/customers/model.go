package customers


type Customer struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Address   string `json:"address"`
}


type CreateCustomerRequest struct {
	Name    string `json:"name" validate:"required"`
	Email   string `json:"email" validate:"required,email"`
	Phone   string `json:"phone" validate:"required"`
	Address string `json:"address" validate:"required"`
}

type UpdateCustomerRequest struct {
	Name    *string `json:"name,omitempty"`
	Email   *string `json:"email,omitempty" validate:"omitempty,email"`
	Phone   *string `json:"phone,omitempty"`
	Address *string `json:"address,omitempty"`
}



