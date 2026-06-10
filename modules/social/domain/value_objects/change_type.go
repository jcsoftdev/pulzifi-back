package valueobjects

// ChangeType classifies a detected change between two consecutive social snapshots.
type ChangeType string

const (
	// ChangeTypeNewPost is emitted when a post appears in the latest snapshot
	// that was not present in the previous snapshot.
	ChangeTypeNewPost ChangeType = "new_post"

	// ChangeTypeRemovedPost is emitted when a post present in the previous
	// snapshot is absent from the latest snapshot.
	ChangeTypeRemovedPost ChangeType = "removed_post"

	// ChangeTypeCaptionEdited is emitted when the caption of an existing post
	// has changed between snapshots.
	ChangeTypeCaptionEdited ChangeType = "caption_edited"

	// ChangeTypeBioChanged is emitted when the profile biography text differs
	// between snapshots.
	ChangeTypeBioChanged ChangeType = "bio_changed"

	// ChangeTypeFollowersChanged is emitted when the follower count delta
	// exceeds the configured threshold (default: any change).
	ChangeTypeFollowersChanged ChangeType = "followers_changed"

	// ChangeTypeAvatarChanged is emitted when the stored avatar URL differs
	// between snapshots (indicating a profile picture change).
	ChangeTypeAvatarChanged ChangeType = "avatar_changed"

	// ChangeTypeDisplayNameChanged is emitted when the display name differs
	// between snapshots.
	ChangeTypeDisplayNameChanged ChangeType = "display_name_changed"
)

// String returns the string representation of the change type.
func (ct ChangeType) String() string {
	return string(ct)
}
