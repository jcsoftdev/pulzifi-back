-- Normalize social_snapshots.data JSONB from PascalCase to snake_case.
-- Snapshots captured before json tags were added stored keys like "MediaURL", "FollowersCount", etc.
-- After this migration all rows use the canonical snake_case keys the Go structs expect.

UPDATE social_snapshots
SET data = jsonb_build_object(
    'bio',              COALESCE(data->>'Bio',             data->>'bio',              ''),
    'display_name',     COALESCE(data->>'DisplayName',     data->>'display_name',     ''),
    'avatar_url',       COALESCE(data->>'AvatarURL',       data->>'avatar_url',       ''),
    'followers_count',  COALESCE((data->>'FollowersCount')::int, (data->>'followers_count')::int, 0),
    'following_count',  COALESCE((data->>'FollowingCount')::int, (data->>'following_count')::int, 0),
    'posts_count',      COALESCE((data->>'PostsCount')::int,     (data->>'posts_count')::int,     0),
    'posts', (
        SELECT COALESCE(jsonb_agg(
            jsonb_build_object(
                'external_id',      COALESCE(p->>'ExternalID',      p->>'external_id',      ''),
                'caption',          COALESCE(p->>'Caption',         p->>'caption',          ''),
                'media_url',        COALESCE(p->>'MediaURL',        p->>'media_url',        ''),
                'stored_media_url', COALESCE(p->>'StoredMediaURL',  p->>'stored_media_url', ''),
                'posted_at',        COALESCE(p->>'PostedAt',        p->>'posted_at',        '0001-01-01T00:00:00Z'),
                'likes_count',      COALESCE((p->>'LikesCount')::int,     (p->>'likes_count')::int,     0),
                'comments_count',   COALESCE((p->>'CommentsCount')::int,  (p->>'comments_count')::int,  0)
            )
        ), '[]'::jsonb)
        FROM jsonb_array_elements(
            COALESCE(data->'Posts', data->'posts', '[]'::jsonb)
        ) AS p
    )
)
WHERE status = 'success'
  AND data IS NOT NULL;
