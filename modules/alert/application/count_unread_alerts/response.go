package countunreadalerts

type CountUnreadAlertsResponse struct {
	HasNotifications  bool `json:"has_notifications"`
	NotificationCount int  `json:"notification_count"`
}
