import { MigrateUpArgs, MigrateDownArgs, sql } from '@payloadcms/db-postgres'

export async function up({ db, payload, req }: MigrateUpArgs): Promise<void> {
  await db.execute(sql`
CREATE TABLE IF NOT EXISTS "cms"."media" (
  "id" serial PRIMARY KEY NOT NULL,
  "alt" varchar,
  "updated_at" timestamp(3) with time zone NOT NULL DEFAULT now(),
  "created_at" timestamp(3) with time zone NOT NULL DEFAULT now(),
  "url" varchar,
  "thumbnail_u_r_l" varchar,
  "filename" varchar,
  "mime_type" varchar,
  "filesize" numeric,
  "width" numeric,
  "height" numeric,
  "focal_x" numeric,
  "focal_y" numeric
);

CREATE TABLE IF NOT EXISTS "cms"."users_sessions" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "created_at" timestamp(3) with time zone,
  "expires_at" timestamp(3) with time zone NOT NULL
);

CREATE TABLE IF NOT EXISTS "cms"."users" (
  "id" serial PRIMARY KEY NOT NULL,
  "updated_at" timestamp(3) with time zone NOT NULL DEFAULT now(),
  "created_at" timestamp(3) with time zone NOT NULL DEFAULT now(),
  "email" varchar NOT NULL,
  "reset_password_token" varchar,
  "reset_password_expiration" timestamp(3) with time zone,
  "salt" varchar,
  "hash" varchar,
  "login_attempts" numeric DEFAULT 0,
  "lock_until" timestamp(3) with time zone
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_hero_dashboard_alerts" (
  "_order" integer NOT NULL,
  "_parent_id" varchar NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "tone" "cms"."enum_pages_blocks_hero_dashboard_alerts_tone" DEFAULT 'signal',
  "icon" varchar,
  "site" varchar,
  "title" varchar,
  "detail" varchar,
  "time" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_hero_kpis" (
  "_order" integer NOT NULL,
  "_parent_id" varchar NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "label" varchar,
  "value" varchar,
  "delta" varchar,
  "delta_direction" "cms"."enum_pages_blocks_hero_kpis_delta_direction" DEFAULT 'up'
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_hero" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "eyebrow_badge" varchar,
  "eyebrow_text" varchar,
  "headline" varchar,
  "headline_highlight" varchar,
  "subheadline" varchar,
  "primary_cta_label" varchar,
  "primary_cta_href" varchar,
  "secondary_cta_label" varchar,
  "secondary_cta_href" varchar,
  "trust_line" varchar,
  "image_id" integer,
  "ai_insight_title" varchar,
  "ai_insight_body" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_logos_items" (
  "_order" integer NOT NULL,
  "_parent_id" varchar NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "text" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_logos" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "label" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_problem_cards" (
  "_order" integer NOT NULL,
  "_parent_id" varchar NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "metric" varchar,
  "label" varchar,
  "description" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_problem" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "eyebrow" varchar,
  "headline" varchar,
  "headline_highlight" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_stats_items" (
  "_order" integer NOT NULL,
  "_parent_id" varchar NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "value" varchar,
  "label" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_stats" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_how_it_works_steps" (
  "_order" integer NOT NULL,
  "_parent_id" varchar NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "step" numeric,
  "icon" varchar,
  "title" varchar,
  "description" varchar,
  "mock_type" "cms"."enum_pages_blocks_how_it_works_steps_mock_type" DEFAULT 'url',
  "mock_text" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_how_it_works" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "eyebrow" varchar,
  "headline" varchar,
  "headline_highlight" varchar,
  "subheadline" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_features_bullets" (
  "_order" integer NOT NULL,
  "_parent_id" varchar NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "title" varchar,
  "description" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_features_demo_actions" (
  "_order" integer NOT NULL,
  "_parent_id" varchar NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "label" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_features_cards" (
  "_order" integer NOT NULL,
  "_parent_id" varchar NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "title" varchar,
  "description" varchar,
  "image_id" integer
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_features" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "eyebrow" varchar,
  "headline" varchar,
  "headline_highlight" varchar,
  "intro" varchar,
  "demo_title" varchar,
  "demo_badge" varchar,
  "demo_site" varchar,
  "demo_change" varchar,
  "demo_analysis" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_ai_intelligence_tabs_items" (
  "_order" integer NOT NULL,
  "_parent_id" varchar NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "title" varchar,
  "body" varchar,
  "image" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_ai_intelligence_tabs" (
  "_order" integer NOT NULL,
  "_parent_id" varchar NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "label" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_ai_intelligence" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "eyebrow" varchar,
  "headline" varchar,
  "headline_highlight" varchar,
  "subheadline" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_insights" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_industries_items" (
  "_order" integer NOT NULL,
  "_parent_id" varchar NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "icon" varchar,
  "title" varchar,
  "description" varchar,
  "real_win" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_industries" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "compact_mode" boolean DEFAULT true,
  "eyebrow" varchar,
  "headline" varchar,
  "headline_highlight" varchar,
  "subheadline" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_comparison_columns" (
  "_order" integer NOT NULL,
  "_parent_id" varchar NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "name" varchar,
  "is_us" boolean
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_comparison_rows_cells" (
  "_order" integer NOT NULL,
  "_parent_id" varchar NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "state" "cms"."enum_pages_blocks_comparison_rows_cells_state" DEFAULT 'yes',
  "note" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_comparison_rows" (
  "_order" integer NOT NULL,
  "_parent_id" varchar NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "feature" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_comparison" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "eyebrow" varchar,
  "headline" varchar,
  "headline_highlight" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_pricing" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "eyebrow" varchar,
  "headline" varchar,
  "headline_highlight" varchar,
  "subheadline" varchar,
  "guarantee_note" varchar,
  "billing_monthly_label" varchar,
  "billing_annual_label" varchar,
  "billing_annual_badge" varchar,
  "billing_annual_note" varchar,
  "compare_plans_headline" varchar,
  "features_label" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_testimonials_items" (
  "_order" integer NOT NULL,
  "_parent_id" varchar NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "quote" varchar,
  "author" varchar,
  "role" varchar,
  "avatar_id" integer
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_testimonials" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "eyebrow" varchar,
  "headline" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_faq_items" (
  "_order" integer NOT NULL,
  "_parent_id" varchar NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "question" varchar,
  "answer" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_faq" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "eyebrow" varchar,
  "headline" varchar,
  "subheadline" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_rich_text" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "content" jsonb,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_cta" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "eyebrow" varchar,
  "headline" varchar,
  "headline_highlight" varchar,
  "subtext" varchar,
  "primary_label" varchar,
  "primary_href" varchar,
  "secondary_label" varchar,
  "secondary_href" varchar,
  "risk_note" varchar,
  "variant" "cms"."enum_pages_blocks_cta_variant" DEFAULT 'primary',
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_image" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "image_id" integer,
  "caption" varchar,
  "size" "cms"."enum_pages_blocks_image_size" DEFAULT 'contained',
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_login_form" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "headline" varchar,
  "subheadline" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_register_form" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "headline" varchar,
  "subheadline" varchar,
  "trial_badge" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages_blocks_block_ref" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "ref_id" integer,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."pages" (
  "id" serial PRIMARY KEY NOT NULL,
  "title" varchar,
  "slug" varchar,
  "meta_title" varchar,
  "meta_description" varchar,
  "meta_image_id" integer,
  "updated_at" timestamp(3) with time zone NOT NULL DEFAULT now(),
  "created_at" timestamp(3) with time zone NOT NULL DEFAULT now(),
  "_status" "cms"."enum_pages_status" DEFAULT 'draft'
);

CREATE TABLE IF NOT EXISTS "cms"."pages_rels" (
  "id" serial PRIMARY KEY NOT NULL,
  "order" integer,
  "parent_id" integer NOT NULL,
  "path" varchar NOT NULL,
  "plans_id" integer
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_hero_dashboard_alerts" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "tone" "cms"."enum__pages_v_blocks_hero_dashboard_alerts_tone" DEFAULT 'signal',
  "icon" varchar,
  "site" varchar,
  "title" varchar,
  "detail" varchar,
  "time" varchar,
  "_uuid" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_hero_kpis" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "label" varchar,
  "value" varchar,
  "delta" varchar,
  "delta_direction" "cms"."enum__pages_v_blocks_hero_kpis_delta_direction" DEFAULT 'up',
  "_uuid" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_hero" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "eyebrow_badge" varchar,
  "eyebrow_text" varchar,
  "headline" varchar,
  "headline_highlight" varchar,
  "subheadline" varchar,
  "primary_cta_label" varchar,
  "primary_cta_href" varchar,
  "secondary_cta_label" varchar,
  "secondary_cta_href" varchar,
  "trust_line" varchar,
  "image_id" integer,
  "ai_insight_title" varchar,
  "ai_insight_body" varchar,
  "_uuid" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_logos_items" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "text" varchar,
  "_uuid" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_logos" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "label" varchar,
  "_uuid" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_problem_cards" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "metric" varchar,
  "label" varchar,
  "description" varchar,
  "_uuid" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_problem" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "eyebrow" varchar,
  "headline" varchar,
  "headline_highlight" varchar,
  "_uuid" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_stats_items" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "value" varchar,
  "label" varchar,
  "_uuid" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_stats" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "_uuid" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_how_it_works_steps" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "step" numeric,
  "icon" varchar,
  "title" varchar,
  "description" varchar,
  "mock_type" "cms"."enum__pages_v_blocks_how_it_works_steps_mock_type" DEFAULT 'url',
  "mock_text" varchar,
  "_uuid" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_how_it_works" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "eyebrow" varchar,
  "headline" varchar,
  "headline_highlight" varchar,
  "subheadline" varchar,
  "_uuid" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_features_bullets" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "title" varchar,
  "description" varchar,
  "_uuid" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_features_demo_actions" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "label" varchar,
  "_uuid" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_features_cards" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "title" varchar,
  "description" varchar,
  "image_id" integer,
  "_uuid" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_features" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "eyebrow" varchar,
  "headline" varchar,
  "headline_highlight" varchar,
  "intro" varchar,
  "demo_title" varchar,
  "demo_badge" varchar,
  "demo_site" varchar,
  "demo_change" varchar,
  "demo_analysis" varchar,
  "_uuid" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_ai_intelligence_tabs_items" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "title" varchar,
  "body" varchar,
  "image" varchar,
  "_uuid" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_ai_intelligence_tabs" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "label" varchar,
  "_uuid" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_ai_intelligence" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "eyebrow" varchar,
  "headline" varchar,
  "headline_highlight" varchar,
  "subheadline" varchar,
  "_uuid" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_insights" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "_uuid" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_industries_items" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "icon" varchar,
  "title" varchar,
  "description" varchar,
  "real_win" varchar,
  "_uuid" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_industries" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "compact_mode" boolean DEFAULT true,
  "eyebrow" varchar,
  "headline" varchar,
  "headline_highlight" varchar,
  "subheadline" varchar,
  "_uuid" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_comparison_columns" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "name" varchar,
  "is_us" boolean,
  "_uuid" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_comparison_rows_cells" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "state" "cms"."enum__pages_v_blocks_comparison_rows_cells_state" DEFAULT 'yes',
  "note" varchar,
  "_uuid" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_comparison_rows" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "feature" varchar,
  "_uuid" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_comparison" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "eyebrow" varchar,
  "headline" varchar,
  "headline_highlight" varchar,
  "_uuid" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_pricing" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "eyebrow" varchar,
  "headline" varchar,
  "headline_highlight" varchar,
  "subheadline" varchar,
  "guarantee_note" varchar,
  "billing_monthly_label" varchar,
  "billing_annual_label" varchar,
  "billing_annual_badge" varchar,
  "billing_annual_note" varchar,
  "compare_plans_headline" varchar,
  "features_label" varchar,
  "_uuid" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_testimonials_items" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "quote" varchar,
  "author" varchar,
  "role" varchar,
  "avatar_id" integer,
  "_uuid" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_testimonials" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "eyebrow" varchar,
  "headline" varchar,
  "_uuid" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_faq_items" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "question" varchar,
  "answer" varchar,
  "_uuid" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_faq" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "eyebrow" varchar,
  "headline" varchar,
  "subheadline" varchar,
  "_uuid" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_rich_text" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "content" jsonb,
  "_uuid" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_cta" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "eyebrow" varchar,
  "headline" varchar,
  "headline_highlight" varchar,
  "subtext" varchar,
  "primary_label" varchar,
  "primary_href" varchar,
  "secondary_label" varchar,
  "secondary_href" varchar,
  "risk_note" varchar,
  "variant" "cms"."enum__pages_v_blocks_cta_variant" DEFAULT 'primary',
  "_uuid" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_image" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "image_id" integer,
  "caption" varchar,
  "size" "cms"."enum__pages_v_blocks_image_size" DEFAULT 'contained',
  "_uuid" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_login_form" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "headline" varchar,
  "subheadline" varchar,
  "_uuid" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_register_form" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "headline" varchar,
  "subheadline" varchar,
  "trial_badge" varchar,
  "_uuid" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_blocks_block_ref" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "_path" text NOT NULL,
  "id" serial PRIMARY KEY NOT NULL,
  "ref_id" integer,
  "_uuid" varchar,
  "block_name" varchar
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v" (
  "id" serial PRIMARY KEY NOT NULL,
  "parent_id" integer,
  "version_title" varchar,
  "version_slug" varchar,
  "version_meta_title" varchar,
  "version_meta_description" varchar,
  "version_meta_image_id" integer,
  "version_updated_at" timestamp(3) with time zone,
  "version_created_at" timestamp(3) with time zone,
  "version__status" "cms"."enum__pages_v_version_status" DEFAULT 'draft',
  "created_at" timestamp(3) with time zone NOT NULL DEFAULT now(),
  "updated_at" timestamp(3) with time zone NOT NULL DEFAULT now(),
  "latest" boolean
);

CREATE TABLE IF NOT EXISTS "cms"."_pages_v_rels" (
  "id" serial PRIMARY KEY NOT NULL,
  "order" integer,
  "parent_id" integer NOT NULL,
  "path" varchar NOT NULL,
  "plans_id" integer
);

CREATE TABLE IF NOT EXISTS "cms"."posts" (
  "id" serial PRIMARY KEY NOT NULL,
  "title" varchar,
  "slug" varchar,
  "hero_image_id" integer,
  "excerpt" varchar,
  "content" jsonb,
  "author" varchar,
  "category" "cms"."enum_posts_category",
  "published_at" timestamp(3) with time zone,
  "meta_title" varchar,
  "meta_description" varchar,
  "meta_image_id" integer,
  "updated_at" timestamp(3) with time zone NOT NULL DEFAULT now(),
  "created_at" timestamp(3) with time zone NOT NULL DEFAULT now(),
  "_status" "cms"."enum_posts_status" DEFAULT 'draft'
);

CREATE TABLE IF NOT EXISTS "cms"."_posts_v" (
  "id" serial PRIMARY KEY NOT NULL,
  "parent_id" integer,
  "version_title" varchar,
  "version_slug" varchar,
  "version_hero_image_id" integer,
  "version_excerpt" varchar,
  "version_content" jsonb,
  "version_author" varchar,
  "version_category" "cms"."enum__posts_v_version_category",
  "version_published_at" timestamp(3) with time zone,
  "version_meta_title" varchar,
  "version_meta_description" varchar,
  "version_meta_image_id" integer,
  "version_updated_at" timestamp(3) with time zone,
  "version_created_at" timestamp(3) with time zone,
  "version__status" "cms"."enum__posts_v_version_status" DEFAULT 'draft',
  "created_at" timestamp(3) with time zone NOT NULL DEFAULT now(),
  "updated_at" timestamp(3) with time zone NOT NULL DEFAULT now(),
  "latest" boolean
);

CREATE TABLE IF NOT EXISTS "cms"."plans_features" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "text" varchar,
  "included" boolean DEFAULT true
);

CREATE TABLE IF NOT EXISTS "cms"."plans" (
  "id" serial PRIMARY KEY NOT NULL,
  "name" varchar NOT NULL,
  "price" varchar NOT NULL,
  "period" varchar,
  "tagline" varchar,
  "price_annual" varchar,
  "cta_label" varchar,
  "cta_href" varchar,
  "highlighted" boolean,
  "popular_badge" varchar,
  "updated_at" timestamp(3) with time zone NOT NULL DEFAULT now(),
  "created_at" timestamp(3) with time zone NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS "cms"."payload_kv" (
  "id" serial PRIMARY KEY NOT NULL,
  "key" varchar NOT NULL,
  "data" jsonb NOT NULL
);

CREATE TABLE IF NOT EXISTS "cms"."payload_locked_documents" (
  "id" serial PRIMARY KEY NOT NULL,
  "global_slug" varchar,
  "updated_at" timestamp(3) with time zone NOT NULL DEFAULT now(),
  "created_at" timestamp(3) with time zone NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS "cms"."payload_locked_documents_rels" (
  "id" serial PRIMARY KEY NOT NULL,
  "order" integer,
  "parent_id" integer NOT NULL,
  "path" varchar NOT NULL,
  "media_id" integer,
  "users_id" integer,
  "block_library_id" integer,
  "pages_id" integer,
  "posts_id" integer,
  "plans_id" integer
);

CREATE TABLE IF NOT EXISTS "cms"."payload_preferences" (
  "id" serial PRIMARY KEY NOT NULL,
  "key" varchar,
  "value" jsonb,
  "updated_at" timestamp(3) with time zone NOT NULL DEFAULT now(),
  "created_at" timestamp(3) with time zone NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS "cms"."payload_preferences_rels" (
  "id" serial PRIMARY KEY NOT NULL,
  "order" integer,
  "parent_id" integer NOT NULL,
  "path" varchar NOT NULL,
  "users_id" integer
);

CREATE TABLE IF NOT EXISTS "cms"."payload_migrations" (
  "id" serial PRIMARY KEY NOT NULL,
  "name" varchar,
  "batch" numeric,
  "updated_at" timestamp(3) with time zone NOT NULL DEFAULT now(),
  "created_at" timestamp(3) with time zone NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS "cms"."navbar_links" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "label" varchar NOT NULL,
  "href" varchar NOT NULL
);

CREATE TABLE IF NOT EXISTS "cms"."navbar" (
  "id" serial PRIMARY KEY NOT NULL,
  "logo_id" integer,
  "signin_label" varchar DEFAULT 'Sign in',
  "signin_href" varchar DEFAULT '/login',
  "primary_cta_label" varchar DEFAULT 'Start Monitoring Free',
  "primary_cta_href" varchar DEFAULT '/register',
  "updated_at" timestamp(3) with time zone,
  "created_at" timestamp(3) with time zone
);

CREATE TABLE IF NOT EXISTS "cms"."footer_groups_links" (
  "_order" integer NOT NULL,
  "_parent_id" varchar NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "label" varchar NOT NULL,
  "href" varchar NOT NULL
);

CREATE TABLE IF NOT EXISTS "cms"."footer_groups" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "heading" varchar NOT NULL
);

CREATE TABLE IF NOT EXISTS "cms"."footer_social_links" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "platform" "cms"."enum_footer_social_links_platform" NOT NULL,
  "href" varchar NOT NULL
);

CREATE TABLE IF NOT EXISTS "cms"."footer" (
  "id" serial PRIMARY KEY NOT NULL,
  "logo_id" integer,
  "tagline" varchar,
  "copyright_text" varchar,
  "updated_at" timestamp(3) with time zone,
  "created_at" timestamp(3) with time zone
);

CREATE TABLE IF NOT EXISTS "cms"."theme" (
  "id" serial PRIMARY KEY NOT NULL,
  "page_bg" varchar,
  "page_bg_alt" varchar,
  "card_bg" varchar,
  "dark_surface" varchar,
  "ink_primary" varchar,
  "ink_secondary" varchar,
  "accent_primary" varchar,
  "accent_muted" varchar,
  "accent_gold" varchar,
  "accent_teal" varchar,
  "border" varchar,
  "border_strong" varchar,
  "updated_at" timestamp(3) with time zone,
  "created_at" timestamp(3) with time zone
);

CREATE TABLE IF NOT EXISTS "cms"."login_page_bullets" (
  "_order" integer NOT NULL,
  "_parent_id" integer NOT NULL,
  "id" varchar PRIMARY KEY NOT NULL,
  "text" varchar NOT NULL
);

CREATE TABLE IF NOT EXISTS "cms"."login_page" (
  "id" serial PRIMARY KEY NOT NULL,
  "headline" varchar DEFAULT 'Know before your competitors move.',
  "subheadline" varchar DEFAULT 'Join 2,500+ teams who react first.',
  "welcome_headline" varchar DEFAULT 'Welcome back',
  "welcome_subline" varchar DEFAULT 'Sign in to your account',
  "updated_at" timestamp(3) with time zone,
  "created_at" timestamp(3) with time zone
);

CREATE TABLE IF NOT EXISTS "cms"."contact_page" (
  "id" serial PRIMARY KEY NOT NULL,
  "headline" varchar DEFAULT 'Get in touch',
  "subheadline" varchar DEFAULT 'We''re here to help. Our team responds within one business day.',
  "email" varchar DEFAULT 'support@pulzifi.com',
  "address" varchar DEFAULT 'Boise, ID',
  "updated_at" timestamp(3) with time zone,
  "created_at" timestamp(3) with time zone
);
  `)
}

export async function down({ db }: MigrateDownArgs): Promise<void> {
  // No-op: tables were missing from initial migration
}
