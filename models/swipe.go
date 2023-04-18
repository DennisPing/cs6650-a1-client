package models

type SwipeRequest struct {
	Swiper  string `json:"swiper"`
	Swipee  string `json:"swipee"`
	Comment string `json:"comment"`
}

type SwipeResponse struct {
	Message string `json:"message"`
}
