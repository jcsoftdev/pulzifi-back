package apify

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
	domainerrors "github.com/jcsoftdev/pulzifi-back/modules/social/domain/errors"
)

// instagramActorItem is the raw shape returned by the Apify Instagram actor
// (apidojo~instagram-scraper). Fields are mapped directly from the actor JSON output.
type instagramActorItem struct {
	Username      string `json:"username"`
	FullName      string `json:"fullName"`
	Biography     string `json:"biography"`
	ProfilePicURL string `json:"profilePicUrl"`
	FollowersCount int   `json:"followersCount"`
	FollowsCount   int   `json:"followsCount"`
	PostsCount     int   `json:"postsCount"`
	LatestPosts    []instagramPost `json:"latestPosts"`
}

// instagramPost represents a single post in the actor's latestPosts array.
type instagramPost struct {
	ID            string `json:"id"`
	Caption       string `json:"caption"`
	DisplayURL    string `json:"displayUrl"`
	Timestamp     string `json:"timestamp"`
	LikesCount    int    `json:"likesCount"`
	CommentsCount int    `json:"commentsCount"`
}

// mapInstagramItems normalises Apify Instagram actor output → entities.ProfileData.
// The actor returns one item per profile; only the first item is used.
func mapInstagramItems(items []json.RawMessage) (*entities.ProfileData, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("%w: instagram actor returned empty dataset", domainerrors.ErrFetchFailed)
	}

	var item instagramActorItem
	if err := json.Unmarshal(items[0], &item); err != nil {
		return nil, fmt.Errorf("%w: unmarshal instagram item: %v", domainerrors.ErrFetchFailed, err)
	}

	posts := make([]entities.Post, 0, len(item.LatestPosts))
	for _, p := range item.LatestPosts {
		post := entities.Post{
			ExternalID:    p.ID,
			Caption:       p.Caption,
			MediaURL:      p.DisplayURL,
			LikesCount:    p.LikesCount,
			CommentsCount: p.CommentsCount,
		}

		// Parse ISO 8601 timestamp produced by the actor.
		if p.Timestamp != "" {
			if t, err := time.Parse(time.RFC3339, p.Timestamp); err == nil {
				post.PostedAt = t.UTC()
			} else if t, err := time.Parse("2006-01-02T15:04:05.000Z", p.Timestamp); err == nil {
				post.PostedAt = t.UTC()
			}
		}

		posts = append(posts, post)
	}

	return &entities.ProfileData{
		Bio:            item.Biography,
		DisplayName:    item.FullName,
		AvatarURL:      item.ProfilePicURL,
		FollowersCount: item.FollowersCount,
		FollowingCount: item.FollowsCount,
		PostsCount:     item.PostsCount,
		Posts:          posts,
	}, nil
}
