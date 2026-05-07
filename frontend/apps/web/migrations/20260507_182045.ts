import { MigrateUpArgs, MigrateDownArgs, sql } from '@payloadcms/db-postgres'

export async function up({ db, payload, req }: MigrateUpArgs): Promise<void> {
  await db.execute(sql`
   CREATE TYPE "cms"."enum_pages_blocks_cta_variant" AS ENUM('primary', 'secondary');
  CREATE TYPE "cms"."enum_pages_blocks_image_size" AS ENUM('full', 'contained');
  CREATE TYPE "cms"."enum_pages_status" AS ENUM('draft', 'published');
  CREATE TYPE "cms"."enum__pages_v_blocks_cta_variant" AS ENUM('primary', 'secondary');
  CREATE TYPE "cms"."enum__pages_v_blocks_image_size" AS ENUM('full', 'contained');
  CREATE TYPE "cms"."enum__pages_v_version_status" AS ENUM('draft', 'published');
  CREATE TYPE "cms"."enum_posts_category" AS ENUM('Product', 'Company', 'Guide');
  CREATE TYPE "cms"."enum_posts_status" AS ENUM('draft', 'published');
  CREATE TYPE "cms"."enum__posts_v_version_category" AS ENUM('Product', 'Company', 'Guide');
  CREATE TYPE "cms"."enum__posts_v_version_status" AS ENUM('draft', 'published');
  CREATE TYPE "cms"."enum_landing_blocks_cta_variant" AS ENUM('primary', 'secondary');
  CREATE TYPE "cms"."enum_landing_blocks_image_size" AS ENUM('full', 'contained');
  CREATE TABLE "cms"."media" (
  	"id" serial PRIMARY KEY NOT NULL,
  	"alt" varchar,
  	"updated_at" timestamp(3) with time zone DEFAULT now() NOT NULL,
  	"created_at" timestamp(3) with time zone DEFAULT now() NOT NULL,
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
  
  CREATE TABLE "cms"."users_sessions" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"created_at" timestamp(3) with time zone,
  	"expires_at" timestamp(3) with time zone NOT NULL
  );
  
  CREATE TABLE "cms"."users" (
  	"id" serial PRIMARY KEY NOT NULL,
  	"updated_at" timestamp(3) with time zone DEFAULT now() NOT NULL,
  	"created_at" timestamp(3) with time zone DEFAULT now() NOT NULL,
  	"email" varchar NOT NULL,
  	"reset_password_token" varchar,
  	"reset_password_expiration" timestamp(3) with time zone,
  	"salt" varchar,
  	"hash" varchar,
  	"login_attempts" numeric DEFAULT 0,
  	"lock_until" timestamp(3) with time zone
  );
  
  CREATE TABLE "cms"."pages_blocks_hero" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"headline" varchar,
  	"subheadline" varchar,
  	"cta_label" varchar,
  	"cta_href" varchar,
  	"image_id" integer,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."pages_blocks_stats_items" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"value" varchar,
  	"label" varchar
  );
  
  CREATE TABLE "cms"."pages_blocks_stats" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."pages_blocks_how_it_works_steps" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"step" numeric,
  	"title" varchar,
  	"description" varchar
  );
  
  CREATE TABLE "cms"."pages_blocks_how_it_works" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."pages_blocks_features_cards" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"title" varchar,
  	"description" varchar,
  	"image_id" integer
  );
  
  CREATE TABLE "cms"."pages_blocks_features" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."pages_blocks_insights" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."pages_blocks_industries" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."pages_blocks_pricing" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"eyebrow" varchar,
  	"headline" varchar,
  	"headline_highlight" varchar,
  	"subheadline" varchar,
  	"guarantee_note" varchar,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."pages_blocks_testimonials_items" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"quote" varchar,
  	"author" varchar,
  	"role" varchar,
  	"avatar_id" integer
  );
  
  CREATE TABLE "cms"."pages_blocks_testimonials" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."pages_blocks_faq_items" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"question" varchar,
  	"answer" varchar
  );
  
  CREATE TABLE "cms"."pages_blocks_faq" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."pages_blocks_rich_text" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"content" jsonb,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."pages_blocks_cta" (
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
  
  CREATE TABLE "cms"."pages_blocks_image" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"image_id" integer,
  	"caption" varchar,
  	"size" "cms"."enum_pages_blocks_image_size" DEFAULT 'contained',
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."pages" (
  	"id" serial PRIMARY KEY NOT NULL,
  	"title" varchar,
  	"slug" varchar,
  	"meta_title" varchar,
  	"meta_description" varchar,
  	"meta_image_id" integer,
  	"updated_at" timestamp(3) with time zone DEFAULT now() NOT NULL,
  	"created_at" timestamp(3) with time zone DEFAULT now() NOT NULL,
  	"_status" "cms"."enum_pages_status" DEFAULT 'draft'
  );
  
  CREATE TABLE "cms"."pages_rels" (
  	"id" serial PRIMARY KEY NOT NULL,
  	"order" integer,
  	"parent_id" integer NOT NULL,
  	"path" varchar NOT NULL,
  	"plans_id" integer
  );
  
  CREATE TABLE "cms"."_pages_v_blocks_hero" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" serial PRIMARY KEY NOT NULL,
  	"headline" varchar,
  	"subheadline" varchar,
  	"cta_label" varchar,
  	"cta_href" varchar,
  	"image_id" integer,
  	"_uuid" varchar,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."_pages_v_blocks_stats_items" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"id" serial PRIMARY KEY NOT NULL,
  	"value" varchar,
  	"label" varchar,
  	"_uuid" varchar
  );
  
  CREATE TABLE "cms"."_pages_v_blocks_stats" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" serial PRIMARY KEY NOT NULL,
  	"_uuid" varchar,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."_pages_v_blocks_how_it_works_steps" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"id" serial PRIMARY KEY NOT NULL,
  	"step" numeric,
  	"title" varchar,
  	"description" varchar,
  	"_uuid" varchar
  );
  
  CREATE TABLE "cms"."_pages_v_blocks_how_it_works" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" serial PRIMARY KEY NOT NULL,
  	"_uuid" varchar,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."_pages_v_blocks_features_cards" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"id" serial PRIMARY KEY NOT NULL,
  	"title" varchar,
  	"description" varchar,
  	"image_id" integer,
  	"_uuid" varchar
  );
  
  CREATE TABLE "cms"."_pages_v_blocks_features" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" serial PRIMARY KEY NOT NULL,
  	"_uuid" varchar,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."_pages_v_blocks_insights" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" serial PRIMARY KEY NOT NULL,
  	"_uuid" varchar,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."_pages_v_blocks_industries" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" serial PRIMARY KEY NOT NULL,
  	"_uuid" varchar,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."_pages_v_blocks_pricing" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" serial PRIMARY KEY NOT NULL,
  	"eyebrow" varchar,
  	"headline" varchar,
  	"headline_highlight" varchar,
  	"subheadline" varchar,
  	"guarantee_note" varchar,
  	"_uuid" varchar,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."_pages_v_blocks_testimonials_items" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"id" serial PRIMARY KEY NOT NULL,
  	"quote" varchar,
  	"author" varchar,
  	"role" varchar,
  	"avatar_id" integer,
  	"_uuid" varchar
  );
  
  CREATE TABLE "cms"."_pages_v_blocks_testimonials" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" serial PRIMARY KEY NOT NULL,
  	"_uuid" varchar,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."_pages_v_blocks_faq_items" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"id" serial PRIMARY KEY NOT NULL,
  	"question" varchar,
  	"answer" varchar,
  	"_uuid" varchar
  );
  
  CREATE TABLE "cms"."_pages_v_blocks_faq" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" serial PRIMARY KEY NOT NULL,
  	"_uuid" varchar,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."_pages_v_blocks_rich_text" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" serial PRIMARY KEY NOT NULL,
  	"content" jsonb,
  	"_uuid" varchar,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."_pages_v_blocks_cta" (
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
  
  CREATE TABLE "cms"."_pages_v_blocks_image" (
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
  
  CREATE TABLE "cms"."_pages_v" (
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
  	"created_at" timestamp(3) with time zone DEFAULT now() NOT NULL,
  	"updated_at" timestamp(3) with time zone DEFAULT now() NOT NULL,
  	"latest" boolean
  );
  
  CREATE TABLE "cms"."_pages_v_rels" (
  	"id" serial PRIMARY KEY NOT NULL,
  	"order" integer,
  	"parent_id" integer NOT NULL,
  	"path" varchar NOT NULL,
  	"plans_id" integer
  );
  
  CREATE TABLE "cms"."plans_features" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"text" varchar NOT NULL,
  	"included" boolean DEFAULT true
  );
  
  CREATE TABLE "cms"."plans" (
  	"id" serial PRIMARY KEY NOT NULL,
  	"name" varchar NOT NULL,
  	"price" varchar NOT NULL,
  	"period" varchar,
  	"tagline" varchar,
  	"cta_label" varchar,
  	"cta_href" varchar,
  	"highlighted" boolean DEFAULT false,
  	"popular_badge" varchar,
  	"updated_at" timestamp(3) with time zone DEFAULT now() NOT NULL,
  	"created_at" timestamp(3) with time zone DEFAULT now() NOT NULL
  );
  
  CREATE TABLE "cms"."posts" (
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
  	"updated_at" timestamp(3) with time zone DEFAULT now() NOT NULL,
  	"created_at" timestamp(3) with time zone DEFAULT now() NOT NULL,
  	"_status" "cms"."enum_posts_status" DEFAULT 'draft'
  );
  
  CREATE TABLE "cms"."_posts_v" (
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
  	"created_at" timestamp(3) with time zone DEFAULT now() NOT NULL,
  	"updated_at" timestamp(3) with time zone DEFAULT now() NOT NULL,
  	"latest" boolean
  );
  
  CREATE TABLE "cms"."payload_kv" (
  	"id" serial PRIMARY KEY NOT NULL,
  	"key" varchar NOT NULL,
  	"data" jsonb NOT NULL
  );
  
  CREATE TABLE "cms"."payload_locked_documents" (
  	"id" serial PRIMARY KEY NOT NULL,
  	"global_slug" varchar,
  	"updated_at" timestamp(3) with time zone DEFAULT now() NOT NULL,
  	"created_at" timestamp(3) with time zone DEFAULT now() NOT NULL
  );
  
  CREATE TABLE "cms"."payload_locked_documents_rels" (
  	"id" serial PRIMARY KEY NOT NULL,
  	"order" integer,
  	"parent_id" integer NOT NULL,
  	"path" varchar NOT NULL,
  	"media_id" integer,
  	"users_id" integer,
  	"pages_id" integer,
  	"plans_id" integer,
  	"posts_id" integer
  );
  
  CREATE TABLE "cms"."payload_preferences" (
  	"id" serial PRIMARY KEY NOT NULL,
  	"key" varchar,
  	"value" jsonb,
  	"updated_at" timestamp(3) with time zone DEFAULT now() NOT NULL,
  	"created_at" timestamp(3) with time zone DEFAULT now() NOT NULL
  );
  
  CREATE TABLE "cms"."payload_preferences_rels" (
  	"id" serial PRIMARY KEY NOT NULL,
  	"order" integer,
  	"parent_id" integer NOT NULL,
  	"path" varchar NOT NULL,
  	"users_id" integer
  );
  
  CREATE TABLE "cms"."payload_migrations" (
  	"id" serial PRIMARY KEY NOT NULL,
  	"name" varchar,
  	"batch" numeric,
  	"updated_at" timestamp(3) with time zone DEFAULT now() NOT NULL,
  	"created_at" timestamp(3) with time zone DEFAULT now() NOT NULL
  );
  
  CREATE TABLE "cms"."landing_blocks_hero" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"headline" varchar NOT NULL,
  	"subheadline" varchar,
  	"cta_label" varchar,
  	"cta_href" varchar,
  	"image_id" integer,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."landing_blocks_stats_items" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"value" varchar NOT NULL,
  	"label" varchar NOT NULL
  );
  
  CREATE TABLE "cms"."landing_blocks_stats" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."landing_blocks_how_it_works_steps" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"step" numeric NOT NULL,
  	"title" varchar NOT NULL,
  	"description" varchar NOT NULL
  );
  
  CREATE TABLE "cms"."landing_blocks_how_it_works" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."landing_blocks_features_cards" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"title" varchar NOT NULL,
  	"description" varchar NOT NULL,
  	"image_id" integer
  );
  
  CREATE TABLE "cms"."landing_blocks_features" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."landing_blocks_insights" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."landing_blocks_industries" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."landing_blocks_pricing" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"eyebrow" varchar,
  	"headline" varchar,
  	"headline_highlight" varchar,
  	"subheadline" varchar,
  	"guarantee_note" varchar,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."landing_blocks_testimonials_items" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"quote" varchar NOT NULL,
  	"author" varchar NOT NULL,
  	"role" varchar,
  	"avatar_id" integer
  );
  
  CREATE TABLE "cms"."landing_blocks_testimonials" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."landing_blocks_faq_items" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"question" varchar NOT NULL,
  	"answer" varchar NOT NULL
  );
  
  CREATE TABLE "cms"."landing_blocks_faq" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."landing_blocks_rich_text" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"content" jsonb NOT NULL,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."landing_blocks_cta" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"eyebrow" varchar,
  	"headline" varchar NOT NULL,
  	"headline_highlight" varchar,
  	"subtext" varchar,
  	"primary_label" varchar,
  	"primary_href" varchar,
  	"secondary_label" varchar,
  	"secondary_href" varchar,
  	"risk_note" varchar,
  	"variant" "cms"."enum_landing_blocks_cta_variant" DEFAULT 'primary',
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."landing_blocks_image" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"image_id" integer NOT NULL,
  	"caption" varchar,
  	"size" "cms"."enum_landing_blocks_image_size" DEFAULT 'contained',
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."landing" (
  	"id" serial PRIMARY KEY NOT NULL,
  	"updated_at" timestamp(3) with time zone,
  	"created_at" timestamp(3) with time zone
  );
  
  CREATE TABLE "cms"."landing_rels" (
  	"id" serial PRIMARY KEY NOT NULL,
  	"order" integer,
  	"parent_id" integer NOT NULL,
  	"path" varchar NOT NULL,
  	"plans_id" integer
  );
  
  CREATE TABLE "cms"."navbar_links" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"label" varchar NOT NULL,
  	"href" varchar NOT NULL
  );
  
  CREATE TABLE "cms"."navbar" (
  	"id" serial PRIMARY KEY NOT NULL,
  	"signin_label" varchar,
  	"signin_href" varchar,
  	"primary_cta_label" varchar,
  	"primary_cta_href" varchar,
  	"logo_id" integer,
  	"updated_at" timestamp(3) with time zone,
  	"created_at" timestamp(3) with time zone
  );
  
  CREATE TABLE "cms"."footer_groups_links" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"label" varchar NOT NULL,
  	"href" varchar NOT NULL
  );
  
  CREATE TABLE "cms"."footer_groups" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"heading" varchar NOT NULL
  );
  
  CREATE TABLE "cms"."footer_social_links" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"platform" varchar NOT NULL,
  	"href" varchar NOT NULL
  );
  
  CREATE TABLE "cms"."footer" (
  	"id" serial PRIMARY KEY NOT NULL,
  	"tagline" varchar,
  	"copyright_text" varchar,
  	"logo_id" integer,
  	"updated_at" timestamp(3) with time zone,
  	"created_at" timestamp(3) with time zone
  );
  
  CREATE TABLE "cms"."pricing_page_faq_items" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"question" varchar NOT NULL,
  	"answer" varchar NOT NULL
  );
  
  CREATE TABLE "cms"."pricing_page" (
  	"id" serial PRIMARY KEY NOT NULL,
  	"header_eyebrow" varchar,
  	"header_headline" varchar,
  	"header_headline_highlight" varchar,
  	"header_subheadline" varchar,
  	"guarantee_note" varchar,
  	"faq_eyebrow" varchar,
  	"faq_headline" varchar,
  	"faq_subheadline" varchar,
  	"updated_at" timestamp(3) with time zone,
  	"created_at" timestamp(3) with time zone
  );
  
  CREATE TABLE "cms"."pricing_page_rels" (
  	"id" serial PRIMARY KEY NOT NULL,
  	"order" integer,
  	"parent_id" integer NOT NULL,
  	"path" varchar NOT NULL,
  	"plans_id" integer
  );
  
  CREATE TABLE "cms"."theme" (
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
  
  ALTER TABLE "cms"."users_sessions" ADD CONSTRAINT "users_sessions_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."users"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."pages_blocks_hero" ADD CONSTRAINT "pages_blocks_hero_image_id_media_id_fk" FOREIGN KEY ("image_id") REFERENCES "cms"."media"("id") ON DELETE set null ON UPDATE no action;
  ALTER TABLE "cms"."pages_blocks_hero" ADD CONSTRAINT "pages_blocks_hero_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."pages"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."pages_blocks_stats_items" ADD CONSTRAINT "pages_blocks_stats_items_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."pages_blocks_stats"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."pages_blocks_stats" ADD CONSTRAINT "pages_blocks_stats_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."pages"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."pages_blocks_how_it_works_steps" ADD CONSTRAINT "pages_blocks_how_it_works_steps_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."pages_blocks_how_it_works"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."pages_blocks_how_it_works" ADD CONSTRAINT "pages_blocks_how_it_works_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."pages"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."pages_blocks_features_cards" ADD CONSTRAINT "pages_blocks_features_cards_image_id_media_id_fk" FOREIGN KEY ("image_id") REFERENCES "cms"."media"("id") ON DELETE set null ON UPDATE no action;
  ALTER TABLE "cms"."pages_blocks_features_cards" ADD CONSTRAINT "pages_blocks_features_cards_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."pages_blocks_features"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."pages_blocks_features" ADD CONSTRAINT "pages_blocks_features_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."pages"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."pages_blocks_insights" ADD CONSTRAINT "pages_blocks_insights_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."pages"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."pages_blocks_industries" ADD CONSTRAINT "pages_blocks_industries_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."pages"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."pages_blocks_pricing" ADD CONSTRAINT "pages_blocks_pricing_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."pages"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."pages_blocks_testimonials_items" ADD CONSTRAINT "pages_blocks_testimonials_items_avatar_id_media_id_fk" FOREIGN KEY ("avatar_id") REFERENCES "cms"."media"("id") ON DELETE set null ON UPDATE no action;
  ALTER TABLE "cms"."pages_blocks_testimonials_items" ADD CONSTRAINT "pages_blocks_testimonials_items_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."pages_blocks_testimonials"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."pages_blocks_testimonials" ADD CONSTRAINT "pages_blocks_testimonials_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."pages"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."pages_blocks_faq_items" ADD CONSTRAINT "pages_blocks_faq_items_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."pages_blocks_faq"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."pages_blocks_faq" ADD CONSTRAINT "pages_blocks_faq_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."pages"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."pages_blocks_rich_text" ADD CONSTRAINT "pages_blocks_rich_text_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."pages"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."pages_blocks_cta" ADD CONSTRAINT "pages_blocks_cta_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."pages"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."pages_blocks_image" ADD CONSTRAINT "pages_blocks_image_image_id_media_id_fk" FOREIGN KEY ("image_id") REFERENCES "cms"."media"("id") ON DELETE set null ON UPDATE no action;
  ALTER TABLE "cms"."pages_blocks_image" ADD CONSTRAINT "pages_blocks_image_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."pages"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."pages" ADD CONSTRAINT "pages_meta_image_id_media_id_fk" FOREIGN KEY ("meta_image_id") REFERENCES "cms"."media"("id") ON DELETE set null ON UPDATE no action;
  ALTER TABLE "cms"."pages_rels" ADD CONSTRAINT "pages_rels_parent_fk" FOREIGN KEY ("parent_id") REFERENCES "cms"."pages"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."pages_rels" ADD CONSTRAINT "pages_rels_plans_fk" FOREIGN KEY ("plans_id") REFERENCES "cms"."plans"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."_pages_v_blocks_hero" ADD CONSTRAINT "_pages_v_blocks_hero_image_id_media_id_fk" FOREIGN KEY ("image_id") REFERENCES "cms"."media"("id") ON DELETE set null ON UPDATE no action;
  ALTER TABLE "cms"."_pages_v_blocks_hero" ADD CONSTRAINT "_pages_v_blocks_hero_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."_pages_v"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."_pages_v_blocks_stats_items" ADD CONSTRAINT "_pages_v_blocks_stats_items_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."_pages_v_blocks_stats"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."_pages_v_blocks_stats" ADD CONSTRAINT "_pages_v_blocks_stats_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."_pages_v"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."_pages_v_blocks_how_it_works_steps" ADD CONSTRAINT "_pages_v_blocks_how_it_works_steps_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."_pages_v_blocks_how_it_works"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."_pages_v_blocks_how_it_works" ADD CONSTRAINT "_pages_v_blocks_how_it_works_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."_pages_v"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."_pages_v_blocks_features_cards" ADD CONSTRAINT "_pages_v_blocks_features_cards_image_id_media_id_fk" FOREIGN KEY ("image_id") REFERENCES "cms"."media"("id") ON DELETE set null ON UPDATE no action;
  ALTER TABLE "cms"."_pages_v_blocks_features_cards" ADD CONSTRAINT "_pages_v_blocks_features_cards_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."_pages_v_blocks_features"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."_pages_v_blocks_features" ADD CONSTRAINT "_pages_v_blocks_features_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."_pages_v"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."_pages_v_blocks_insights" ADD CONSTRAINT "_pages_v_blocks_insights_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."_pages_v"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."_pages_v_blocks_industries" ADD CONSTRAINT "_pages_v_blocks_industries_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."_pages_v"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."_pages_v_blocks_pricing" ADD CONSTRAINT "_pages_v_blocks_pricing_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."_pages_v"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."_pages_v_blocks_testimonials_items" ADD CONSTRAINT "_pages_v_blocks_testimonials_items_avatar_id_media_id_fk" FOREIGN KEY ("avatar_id") REFERENCES "cms"."media"("id") ON DELETE set null ON UPDATE no action;
  ALTER TABLE "cms"."_pages_v_blocks_testimonials_items" ADD CONSTRAINT "_pages_v_blocks_testimonials_items_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."_pages_v_blocks_testimonials"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."_pages_v_blocks_testimonials" ADD CONSTRAINT "_pages_v_blocks_testimonials_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."_pages_v"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."_pages_v_blocks_faq_items" ADD CONSTRAINT "_pages_v_blocks_faq_items_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."_pages_v_blocks_faq"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."_pages_v_blocks_faq" ADD CONSTRAINT "_pages_v_blocks_faq_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."_pages_v"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."_pages_v_blocks_rich_text" ADD CONSTRAINT "_pages_v_blocks_rich_text_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."_pages_v"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."_pages_v_blocks_cta" ADD CONSTRAINT "_pages_v_blocks_cta_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."_pages_v"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."_pages_v_blocks_image" ADD CONSTRAINT "_pages_v_blocks_image_image_id_media_id_fk" FOREIGN KEY ("image_id") REFERENCES "cms"."media"("id") ON DELETE set null ON UPDATE no action;
  ALTER TABLE "cms"."_pages_v_blocks_image" ADD CONSTRAINT "_pages_v_blocks_image_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."_pages_v"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."_pages_v" ADD CONSTRAINT "_pages_v_parent_id_pages_id_fk" FOREIGN KEY ("parent_id") REFERENCES "cms"."pages"("id") ON DELETE set null ON UPDATE no action;
  ALTER TABLE "cms"."_pages_v" ADD CONSTRAINT "_pages_v_version_meta_image_id_media_id_fk" FOREIGN KEY ("version_meta_image_id") REFERENCES "cms"."media"("id") ON DELETE set null ON UPDATE no action;
  ALTER TABLE "cms"."_pages_v_rels" ADD CONSTRAINT "_pages_v_rels_parent_fk" FOREIGN KEY ("parent_id") REFERENCES "cms"."_pages_v"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."_pages_v_rels" ADD CONSTRAINT "_pages_v_rels_plans_fk" FOREIGN KEY ("plans_id") REFERENCES "cms"."plans"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."plans_features" ADD CONSTRAINT "plans_features_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."plans"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."posts" ADD CONSTRAINT "posts_hero_image_id_media_id_fk" FOREIGN KEY ("hero_image_id") REFERENCES "cms"."media"("id") ON DELETE set null ON UPDATE no action;
  ALTER TABLE "cms"."posts" ADD CONSTRAINT "posts_meta_image_id_media_id_fk" FOREIGN KEY ("meta_image_id") REFERENCES "cms"."media"("id") ON DELETE set null ON UPDATE no action;
  ALTER TABLE "cms"."_posts_v" ADD CONSTRAINT "_posts_v_parent_id_posts_id_fk" FOREIGN KEY ("parent_id") REFERENCES "cms"."posts"("id") ON DELETE set null ON UPDATE no action;
  ALTER TABLE "cms"."_posts_v" ADD CONSTRAINT "_posts_v_version_hero_image_id_media_id_fk" FOREIGN KEY ("version_hero_image_id") REFERENCES "cms"."media"("id") ON DELETE set null ON UPDATE no action;
  ALTER TABLE "cms"."_posts_v" ADD CONSTRAINT "_posts_v_version_meta_image_id_media_id_fk" FOREIGN KEY ("version_meta_image_id") REFERENCES "cms"."media"("id") ON DELETE set null ON UPDATE no action;
  ALTER TABLE "cms"."payload_locked_documents_rels" ADD CONSTRAINT "payload_locked_documents_rels_parent_fk" FOREIGN KEY ("parent_id") REFERENCES "cms"."payload_locked_documents"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."payload_locked_documents_rels" ADD CONSTRAINT "payload_locked_documents_rels_media_fk" FOREIGN KEY ("media_id") REFERENCES "cms"."media"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."payload_locked_documents_rels" ADD CONSTRAINT "payload_locked_documents_rels_users_fk" FOREIGN KEY ("users_id") REFERENCES "cms"."users"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."payload_locked_documents_rels" ADD CONSTRAINT "payload_locked_documents_rels_pages_fk" FOREIGN KEY ("pages_id") REFERENCES "cms"."pages"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."payload_locked_documents_rels" ADD CONSTRAINT "payload_locked_documents_rels_plans_fk" FOREIGN KEY ("plans_id") REFERENCES "cms"."plans"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."payload_locked_documents_rels" ADD CONSTRAINT "payload_locked_documents_rels_posts_fk" FOREIGN KEY ("posts_id") REFERENCES "cms"."posts"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."payload_preferences_rels" ADD CONSTRAINT "payload_preferences_rels_parent_fk" FOREIGN KEY ("parent_id") REFERENCES "cms"."payload_preferences"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."payload_preferences_rels" ADD CONSTRAINT "payload_preferences_rels_users_fk" FOREIGN KEY ("users_id") REFERENCES "cms"."users"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."landing_blocks_hero" ADD CONSTRAINT "landing_blocks_hero_image_id_media_id_fk" FOREIGN KEY ("image_id") REFERENCES "cms"."media"("id") ON DELETE set null ON UPDATE no action;
  ALTER TABLE "cms"."landing_blocks_hero" ADD CONSTRAINT "landing_blocks_hero_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."landing"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."landing_blocks_stats_items" ADD CONSTRAINT "landing_blocks_stats_items_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."landing_blocks_stats"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."landing_blocks_stats" ADD CONSTRAINT "landing_blocks_stats_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."landing"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."landing_blocks_how_it_works_steps" ADD CONSTRAINT "landing_blocks_how_it_works_steps_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."landing_blocks_how_it_works"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."landing_blocks_how_it_works" ADD CONSTRAINT "landing_blocks_how_it_works_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."landing"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."landing_blocks_features_cards" ADD CONSTRAINT "landing_blocks_features_cards_image_id_media_id_fk" FOREIGN KEY ("image_id") REFERENCES "cms"."media"("id") ON DELETE set null ON UPDATE no action;
  ALTER TABLE "cms"."landing_blocks_features_cards" ADD CONSTRAINT "landing_blocks_features_cards_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."landing_blocks_features"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."landing_blocks_features" ADD CONSTRAINT "landing_blocks_features_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."landing"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."landing_blocks_insights" ADD CONSTRAINT "landing_blocks_insights_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."landing"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."landing_blocks_industries" ADD CONSTRAINT "landing_blocks_industries_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."landing"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."landing_blocks_pricing" ADD CONSTRAINT "landing_blocks_pricing_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."landing"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."landing_blocks_testimonials_items" ADD CONSTRAINT "landing_blocks_testimonials_items_avatar_id_media_id_fk" FOREIGN KEY ("avatar_id") REFERENCES "cms"."media"("id") ON DELETE set null ON UPDATE no action;
  ALTER TABLE "cms"."landing_blocks_testimonials_items" ADD CONSTRAINT "landing_blocks_testimonials_items_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."landing_blocks_testimonials"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."landing_blocks_testimonials" ADD CONSTRAINT "landing_blocks_testimonials_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."landing"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."landing_blocks_faq_items" ADD CONSTRAINT "landing_blocks_faq_items_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."landing_blocks_faq"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."landing_blocks_faq" ADD CONSTRAINT "landing_blocks_faq_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."landing"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."landing_blocks_rich_text" ADD CONSTRAINT "landing_blocks_rich_text_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."landing"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."landing_blocks_cta" ADD CONSTRAINT "landing_blocks_cta_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."landing"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."landing_blocks_image" ADD CONSTRAINT "landing_blocks_image_image_id_media_id_fk" FOREIGN KEY ("image_id") REFERENCES "cms"."media"("id") ON DELETE set null ON UPDATE no action;
  ALTER TABLE "cms"."landing_blocks_image" ADD CONSTRAINT "landing_blocks_image_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."landing"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."landing_rels" ADD CONSTRAINT "landing_rels_parent_fk" FOREIGN KEY ("parent_id") REFERENCES "cms"."landing"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."landing_rels" ADD CONSTRAINT "landing_rels_plans_fk" FOREIGN KEY ("plans_id") REFERENCES "cms"."plans"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."navbar_links" ADD CONSTRAINT "navbar_links_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."navbar"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."navbar" ADD CONSTRAINT "navbar_logo_id_media_id_fk" FOREIGN KEY ("logo_id") REFERENCES "cms"."media"("id") ON DELETE set null ON UPDATE no action;
  ALTER TABLE "cms"."footer_groups_links" ADD CONSTRAINT "footer_groups_links_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."footer_groups"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."footer_groups" ADD CONSTRAINT "footer_groups_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."footer"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."footer_social_links" ADD CONSTRAINT "footer_social_links_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."footer"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."footer" ADD CONSTRAINT "footer_logo_id_media_id_fk" FOREIGN KEY ("logo_id") REFERENCES "cms"."media"("id") ON DELETE set null ON UPDATE no action;
  ALTER TABLE "cms"."pricing_page_faq_items" ADD CONSTRAINT "pricing_page_faq_items_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."pricing_page"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."pricing_page_rels" ADD CONSTRAINT "pricing_page_rels_parent_fk" FOREIGN KEY ("parent_id") REFERENCES "cms"."pricing_page"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."pricing_page_rels" ADD CONSTRAINT "pricing_page_rels_plans_fk" FOREIGN KEY ("plans_id") REFERENCES "cms"."plans"("id") ON DELETE cascade ON UPDATE no action;
  CREATE INDEX "media_updated_at_idx" ON "cms"."media" USING btree ("updated_at");
  CREATE INDEX "media_created_at_idx" ON "cms"."media" USING btree ("created_at");
  CREATE UNIQUE INDEX "media_filename_idx" ON "cms"."media" USING btree ("filename");
  CREATE INDEX "users_sessions_order_idx" ON "cms"."users_sessions" USING btree ("_order");
  CREATE INDEX "users_sessions_parent_id_idx" ON "cms"."users_sessions" USING btree ("_parent_id");
  CREATE INDEX "users_updated_at_idx" ON "cms"."users" USING btree ("updated_at");
  CREATE INDEX "users_created_at_idx" ON "cms"."users" USING btree ("created_at");
  CREATE UNIQUE INDEX "users_email_idx" ON "cms"."users" USING btree ("email");
  CREATE INDEX "pages_blocks_hero_order_idx" ON "cms"."pages_blocks_hero" USING btree ("_order");
  CREATE INDEX "pages_blocks_hero_parent_id_idx" ON "cms"."pages_blocks_hero" USING btree ("_parent_id");
  CREATE INDEX "pages_blocks_hero_path_idx" ON "cms"."pages_blocks_hero" USING btree ("_path");
  CREATE INDEX "pages_blocks_hero_image_idx" ON "cms"."pages_blocks_hero" USING btree ("image_id");
  CREATE INDEX "pages_blocks_stats_items_order_idx" ON "cms"."pages_blocks_stats_items" USING btree ("_order");
  CREATE INDEX "pages_blocks_stats_items_parent_id_idx" ON "cms"."pages_blocks_stats_items" USING btree ("_parent_id");
  CREATE INDEX "pages_blocks_stats_order_idx" ON "cms"."pages_blocks_stats" USING btree ("_order");
  CREATE INDEX "pages_blocks_stats_parent_id_idx" ON "cms"."pages_blocks_stats" USING btree ("_parent_id");
  CREATE INDEX "pages_blocks_stats_path_idx" ON "cms"."pages_blocks_stats" USING btree ("_path");
  CREATE INDEX "pages_blocks_how_it_works_steps_order_idx" ON "cms"."pages_blocks_how_it_works_steps" USING btree ("_order");
  CREATE INDEX "pages_blocks_how_it_works_steps_parent_id_idx" ON "cms"."pages_blocks_how_it_works_steps" USING btree ("_parent_id");
  CREATE INDEX "pages_blocks_how_it_works_order_idx" ON "cms"."pages_blocks_how_it_works" USING btree ("_order");
  CREATE INDEX "pages_blocks_how_it_works_parent_id_idx" ON "cms"."pages_blocks_how_it_works" USING btree ("_parent_id");
  CREATE INDEX "pages_blocks_how_it_works_path_idx" ON "cms"."pages_blocks_how_it_works" USING btree ("_path");
  CREATE INDEX "pages_blocks_features_cards_order_idx" ON "cms"."pages_blocks_features_cards" USING btree ("_order");
  CREATE INDEX "pages_blocks_features_cards_parent_id_idx" ON "cms"."pages_blocks_features_cards" USING btree ("_parent_id");
  CREATE INDEX "pages_blocks_features_cards_image_idx" ON "cms"."pages_blocks_features_cards" USING btree ("image_id");
  CREATE INDEX "pages_blocks_features_order_idx" ON "cms"."pages_blocks_features" USING btree ("_order");
  CREATE INDEX "pages_blocks_features_parent_id_idx" ON "cms"."pages_blocks_features" USING btree ("_parent_id");
  CREATE INDEX "pages_blocks_features_path_idx" ON "cms"."pages_blocks_features" USING btree ("_path");
  CREATE INDEX "pages_blocks_insights_order_idx" ON "cms"."pages_blocks_insights" USING btree ("_order");
  CREATE INDEX "pages_blocks_insights_parent_id_idx" ON "cms"."pages_blocks_insights" USING btree ("_parent_id");
  CREATE INDEX "pages_blocks_insights_path_idx" ON "cms"."pages_blocks_insights" USING btree ("_path");
  CREATE INDEX "pages_blocks_industries_order_idx" ON "cms"."pages_blocks_industries" USING btree ("_order");
  CREATE INDEX "pages_blocks_industries_parent_id_idx" ON "cms"."pages_blocks_industries" USING btree ("_parent_id");
  CREATE INDEX "pages_blocks_industries_path_idx" ON "cms"."pages_blocks_industries" USING btree ("_path");
  CREATE INDEX "pages_blocks_pricing_order_idx" ON "cms"."pages_blocks_pricing" USING btree ("_order");
  CREATE INDEX "pages_blocks_pricing_parent_id_idx" ON "cms"."pages_blocks_pricing" USING btree ("_parent_id");
  CREATE INDEX "pages_blocks_pricing_path_idx" ON "cms"."pages_blocks_pricing" USING btree ("_path");
  CREATE INDEX "pages_blocks_testimonials_items_order_idx" ON "cms"."pages_blocks_testimonials_items" USING btree ("_order");
  CREATE INDEX "pages_blocks_testimonials_items_parent_id_idx" ON "cms"."pages_blocks_testimonials_items" USING btree ("_parent_id");
  CREATE INDEX "pages_blocks_testimonials_items_avatar_idx" ON "cms"."pages_blocks_testimonials_items" USING btree ("avatar_id");
  CREATE INDEX "pages_blocks_testimonials_order_idx" ON "cms"."pages_blocks_testimonials" USING btree ("_order");
  CREATE INDEX "pages_blocks_testimonials_parent_id_idx" ON "cms"."pages_blocks_testimonials" USING btree ("_parent_id");
  CREATE INDEX "pages_blocks_testimonials_path_idx" ON "cms"."pages_blocks_testimonials" USING btree ("_path");
  CREATE INDEX "pages_blocks_faq_items_order_idx" ON "cms"."pages_blocks_faq_items" USING btree ("_order");
  CREATE INDEX "pages_blocks_faq_items_parent_id_idx" ON "cms"."pages_blocks_faq_items" USING btree ("_parent_id");
  CREATE INDEX "pages_blocks_faq_order_idx" ON "cms"."pages_blocks_faq" USING btree ("_order");
  CREATE INDEX "pages_blocks_faq_parent_id_idx" ON "cms"."pages_blocks_faq" USING btree ("_parent_id");
  CREATE INDEX "pages_blocks_faq_path_idx" ON "cms"."pages_blocks_faq" USING btree ("_path");
  CREATE INDEX "pages_blocks_rich_text_order_idx" ON "cms"."pages_blocks_rich_text" USING btree ("_order");
  CREATE INDEX "pages_blocks_rich_text_parent_id_idx" ON "cms"."pages_blocks_rich_text" USING btree ("_parent_id");
  CREATE INDEX "pages_blocks_rich_text_path_idx" ON "cms"."pages_blocks_rich_text" USING btree ("_path");
  CREATE INDEX "pages_blocks_cta_order_idx" ON "cms"."pages_blocks_cta" USING btree ("_order");
  CREATE INDEX "pages_blocks_cta_parent_id_idx" ON "cms"."pages_blocks_cta" USING btree ("_parent_id");
  CREATE INDEX "pages_blocks_cta_path_idx" ON "cms"."pages_blocks_cta" USING btree ("_path");
  CREATE INDEX "pages_blocks_image_order_idx" ON "cms"."pages_blocks_image" USING btree ("_order");
  CREATE INDEX "pages_blocks_image_parent_id_idx" ON "cms"."pages_blocks_image" USING btree ("_parent_id");
  CREATE INDEX "pages_blocks_image_path_idx" ON "cms"."pages_blocks_image" USING btree ("_path");
  CREATE INDEX "pages_blocks_image_image_idx" ON "cms"."pages_blocks_image" USING btree ("image_id");
  CREATE UNIQUE INDEX "pages_slug_idx" ON "cms"."pages" USING btree ("slug");
  CREATE INDEX "pages_meta_meta_image_idx" ON "cms"."pages" USING btree ("meta_image_id");
  CREATE INDEX "pages_updated_at_idx" ON "cms"."pages" USING btree ("updated_at");
  CREATE INDEX "pages_created_at_idx" ON "cms"."pages" USING btree ("created_at");
  CREATE INDEX "pages__status_idx" ON "cms"."pages" USING btree ("_status");
  CREATE INDEX "pages_rels_order_idx" ON "cms"."pages_rels" USING btree ("order");
  CREATE INDEX "pages_rels_parent_idx" ON "cms"."pages_rels" USING btree ("parent_id");
  CREATE INDEX "pages_rels_path_idx" ON "cms"."pages_rels" USING btree ("path");
  CREATE INDEX "pages_rels_plans_id_idx" ON "cms"."pages_rels" USING btree ("plans_id");
  CREATE INDEX "_pages_v_blocks_hero_order_idx" ON "cms"."_pages_v_blocks_hero" USING btree ("_order");
  CREATE INDEX "_pages_v_blocks_hero_parent_id_idx" ON "cms"."_pages_v_blocks_hero" USING btree ("_parent_id");
  CREATE INDEX "_pages_v_blocks_hero_path_idx" ON "cms"."_pages_v_blocks_hero" USING btree ("_path");
  CREATE INDEX "_pages_v_blocks_hero_image_idx" ON "cms"."_pages_v_blocks_hero" USING btree ("image_id");
  CREATE INDEX "_pages_v_blocks_stats_items_order_idx" ON "cms"."_pages_v_blocks_stats_items" USING btree ("_order");
  CREATE INDEX "_pages_v_blocks_stats_items_parent_id_idx" ON "cms"."_pages_v_blocks_stats_items" USING btree ("_parent_id");
  CREATE INDEX "_pages_v_blocks_stats_order_idx" ON "cms"."_pages_v_blocks_stats" USING btree ("_order");
  CREATE INDEX "_pages_v_blocks_stats_parent_id_idx" ON "cms"."_pages_v_blocks_stats" USING btree ("_parent_id");
  CREATE INDEX "_pages_v_blocks_stats_path_idx" ON "cms"."_pages_v_blocks_stats" USING btree ("_path");
  CREATE INDEX "_pages_v_blocks_how_it_works_steps_order_idx" ON "cms"."_pages_v_blocks_how_it_works_steps" USING btree ("_order");
  CREATE INDEX "_pages_v_blocks_how_it_works_steps_parent_id_idx" ON "cms"."_pages_v_blocks_how_it_works_steps" USING btree ("_parent_id");
  CREATE INDEX "_pages_v_blocks_how_it_works_order_idx" ON "cms"."_pages_v_blocks_how_it_works" USING btree ("_order");
  CREATE INDEX "_pages_v_blocks_how_it_works_parent_id_idx" ON "cms"."_pages_v_blocks_how_it_works" USING btree ("_parent_id");
  CREATE INDEX "_pages_v_blocks_how_it_works_path_idx" ON "cms"."_pages_v_blocks_how_it_works" USING btree ("_path");
  CREATE INDEX "_pages_v_blocks_features_cards_order_idx" ON "cms"."_pages_v_blocks_features_cards" USING btree ("_order");
  CREATE INDEX "_pages_v_blocks_features_cards_parent_id_idx" ON "cms"."_pages_v_blocks_features_cards" USING btree ("_parent_id");
  CREATE INDEX "_pages_v_blocks_features_cards_image_idx" ON "cms"."_pages_v_blocks_features_cards" USING btree ("image_id");
  CREATE INDEX "_pages_v_blocks_features_order_idx" ON "cms"."_pages_v_blocks_features" USING btree ("_order");
  CREATE INDEX "_pages_v_blocks_features_parent_id_idx" ON "cms"."_pages_v_blocks_features" USING btree ("_parent_id");
  CREATE INDEX "_pages_v_blocks_features_path_idx" ON "cms"."_pages_v_blocks_features" USING btree ("_path");
  CREATE INDEX "_pages_v_blocks_insights_order_idx" ON "cms"."_pages_v_blocks_insights" USING btree ("_order");
  CREATE INDEX "_pages_v_blocks_insights_parent_id_idx" ON "cms"."_pages_v_blocks_insights" USING btree ("_parent_id");
  CREATE INDEX "_pages_v_blocks_insights_path_idx" ON "cms"."_pages_v_blocks_insights" USING btree ("_path");
  CREATE INDEX "_pages_v_blocks_industries_order_idx" ON "cms"."_pages_v_blocks_industries" USING btree ("_order");
  CREATE INDEX "_pages_v_blocks_industries_parent_id_idx" ON "cms"."_pages_v_blocks_industries" USING btree ("_parent_id");
  CREATE INDEX "_pages_v_blocks_industries_path_idx" ON "cms"."_pages_v_blocks_industries" USING btree ("_path");
  CREATE INDEX "_pages_v_blocks_pricing_order_idx" ON "cms"."_pages_v_blocks_pricing" USING btree ("_order");
  CREATE INDEX "_pages_v_blocks_pricing_parent_id_idx" ON "cms"."_pages_v_blocks_pricing" USING btree ("_parent_id");
  CREATE INDEX "_pages_v_blocks_pricing_path_idx" ON "cms"."_pages_v_blocks_pricing" USING btree ("_path");
  CREATE INDEX "_pages_v_blocks_testimonials_items_order_idx" ON "cms"."_pages_v_blocks_testimonials_items" USING btree ("_order");
  CREATE INDEX "_pages_v_blocks_testimonials_items_parent_id_idx" ON "cms"."_pages_v_blocks_testimonials_items" USING btree ("_parent_id");
  CREATE INDEX "_pages_v_blocks_testimonials_items_avatar_idx" ON "cms"."_pages_v_blocks_testimonials_items" USING btree ("avatar_id");
  CREATE INDEX "_pages_v_blocks_testimonials_order_idx" ON "cms"."_pages_v_blocks_testimonials" USING btree ("_order");
  CREATE INDEX "_pages_v_blocks_testimonials_parent_id_idx" ON "cms"."_pages_v_blocks_testimonials" USING btree ("_parent_id");
  CREATE INDEX "_pages_v_blocks_testimonials_path_idx" ON "cms"."_pages_v_blocks_testimonials" USING btree ("_path");
  CREATE INDEX "_pages_v_blocks_faq_items_order_idx" ON "cms"."_pages_v_blocks_faq_items" USING btree ("_order");
  CREATE INDEX "_pages_v_blocks_faq_items_parent_id_idx" ON "cms"."_pages_v_blocks_faq_items" USING btree ("_parent_id");
  CREATE INDEX "_pages_v_blocks_faq_order_idx" ON "cms"."_pages_v_blocks_faq" USING btree ("_order");
  CREATE INDEX "_pages_v_blocks_faq_parent_id_idx" ON "cms"."_pages_v_blocks_faq" USING btree ("_parent_id");
  CREATE INDEX "_pages_v_blocks_faq_path_idx" ON "cms"."_pages_v_blocks_faq" USING btree ("_path");
  CREATE INDEX "_pages_v_blocks_rich_text_order_idx" ON "cms"."_pages_v_blocks_rich_text" USING btree ("_order");
  CREATE INDEX "_pages_v_blocks_rich_text_parent_id_idx" ON "cms"."_pages_v_blocks_rich_text" USING btree ("_parent_id");
  CREATE INDEX "_pages_v_blocks_rich_text_path_idx" ON "cms"."_pages_v_blocks_rich_text" USING btree ("_path");
  CREATE INDEX "_pages_v_blocks_cta_order_idx" ON "cms"."_pages_v_blocks_cta" USING btree ("_order");
  CREATE INDEX "_pages_v_blocks_cta_parent_id_idx" ON "cms"."_pages_v_blocks_cta" USING btree ("_parent_id");
  CREATE INDEX "_pages_v_blocks_cta_path_idx" ON "cms"."_pages_v_blocks_cta" USING btree ("_path");
  CREATE INDEX "_pages_v_blocks_image_order_idx" ON "cms"."_pages_v_blocks_image" USING btree ("_order");
  CREATE INDEX "_pages_v_blocks_image_parent_id_idx" ON "cms"."_pages_v_blocks_image" USING btree ("_parent_id");
  CREATE INDEX "_pages_v_blocks_image_path_idx" ON "cms"."_pages_v_blocks_image" USING btree ("_path");
  CREATE INDEX "_pages_v_blocks_image_image_idx" ON "cms"."_pages_v_blocks_image" USING btree ("image_id");
  CREATE INDEX "_pages_v_parent_idx" ON "cms"."_pages_v" USING btree ("parent_id");
  CREATE INDEX "_pages_v_version_version_slug_idx" ON "cms"."_pages_v" USING btree ("version_slug");
  CREATE INDEX "_pages_v_version_meta_version_meta_image_idx" ON "cms"."_pages_v" USING btree ("version_meta_image_id");
  CREATE INDEX "_pages_v_version_version_updated_at_idx" ON "cms"."_pages_v" USING btree ("version_updated_at");
  CREATE INDEX "_pages_v_version_version_created_at_idx" ON "cms"."_pages_v" USING btree ("version_created_at");
  CREATE INDEX "_pages_v_version_version__status_idx" ON "cms"."_pages_v" USING btree ("version__status");
  CREATE INDEX "_pages_v_created_at_idx" ON "cms"."_pages_v" USING btree ("created_at");
  CREATE INDEX "_pages_v_updated_at_idx" ON "cms"."_pages_v" USING btree ("updated_at");
  CREATE INDEX "_pages_v_latest_idx" ON "cms"."_pages_v" USING btree ("latest");
  CREATE INDEX "_pages_v_rels_order_idx" ON "cms"."_pages_v_rels" USING btree ("order");
  CREATE INDEX "_pages_v_rels_parent_idx" ON "cms"."_pages_v_rels" USING btree ("parent_id");
  CREATE INDEX "_pages_v_rels_path_idx" ON "cms"."_pages_v_rels" USING btree ("path");
  CREATE INDEX "_pages_v_rels_plans_id_idx" ON "cms"."_pages_v_rels" USING btree ("plans_id");
  CREATE INDEX "plans_features_order_idx" ON "cms"."plans_features" USING btree ("_order");
  CREATE INDEX "plans_features_parent_id_idx" ON "cms"."plans_features" USING btree ("_parent_id");
  CREATE INDEX "plans_updated_at_idx" ON "cms"."plans" USING btree ("updated_at");
  CREATE INDEX "plans_created_at_idx" ON "cms"."plans" USING btree ("created_at");
  CREATE UNIQUE INDEX "posts_slug_idx" ON "cms"."posts" USING btree ("slug");
  CREATE INDEX "posts_hero_image_idx" ON "cms"."posts" USING btree ("hero_image_id");
  CREATE INDEX "posts_meta_meta_image_idx" ON "cms"."posts" USING btree ("meta_image_id");
  CREATE INDEX "posts_updated_at_idx" ON "cms"."posts" USING btree ("updated_at");
  CREATE INDEX "posts_created_at_idx" ON "cms"."posts" USING btree ("created_at");
  CREATE INDEX "posts__status_idx" ON "cms"."posts" USING btree ("_status");
  CREATE INDEX "_posts_v_parent_idx" ON "cms"."_posts_v" USING btree ("parent_id");
  CREATE INDEX "_posts_v_version_version_slug_idx" ON "cms"."_posts_v" USING btree ("version_slug");
  CREATE INDEX "_posts_v_version_version_hero_image_idx" ON "cms"."_posts_v" USING btree ("version_hero_image_id");
  CREATE INDEX "_posts_v_version_meta_version_meta_image_idx" ON "cms"."_posts_v" USING btree ("version_meta_image_id");
  CREATE INDEX "_posts_v_version_version_updated_at_idx" ON "cms"."_posts_v" USING btree ("version_updated_at");
  CREATE INDEX "_posts_v_version_version_created_at_idx" ON "cms"."_posts_v" USING btree ("version_created_at");
  CREATE INDEX "_posts_v_version_version__status_idx" ON "cms"."_posts_v" USING btree ("version__status");
  CREATE INDEX "_posts_v_created_at_idx" ON "cms"."_posts_v" USING btree ("created_at");
  CREATE INDEX "_posts_v_updated_at_idx" ON "cms"."_posts_v" USING btree ("updated_at");
  CREATE INDEX "_posts_v_latest_idx" ON "cms"."_posts_v" USING btree ("latest");
  CREATE UNIQUE INDEX "payload_kv_key_idx" ON "cms"."payload_kv" USING btree ("key");
  CREATE INDEX "payload_locked_documents_global_slug_idx" ON "cms"."payload_locked_documents" USING btree ("global_slug");
  CREATE INDEX "payload_locked_documents_updated_at_idx" ON "cms"."payload_locked_documents" USING btree ("updated_at");
  CREATE INDEX "payload_locked_documents_created_at_idx" ON "cms"."payload_locked_documents" USING btree ("created_at");
  CREATE INDEX "payload_locked_documents_rels_order_idx" ON "cms"."payload_locked_documents_rels" USING btree ("order");
  CREATE INDEX "payload_locked_documents_rels_parent_idx" ON "cms"."payload_locked_documents_rels" USING btree ("parent_id");
  CREATE INDEX "payload_locked_documents_rels_path_idx" ON "cms"."payload_locked_documents_rels" USING btree ("path");
  CREATE INDEX "payload_locked_documents_rels_media_id_idx" ON "cms"."payload_locked_documents_rels" USING btree ("media_id");
  CREATE INDEX "payload_locked_documents_rels_users_id_idx" ON "cms"."payload_locked_documents_rels" USING btree ("users_id");
  CREATE INDEX "payload_locked_documents_rels_pages_id_idx" ON "cms"."payload_locked_documents_rels" USING btree ("pages_id");
  CREATE INDEX "payload_locked_documents_rels_plans_id_idx" ON "cms"."payload_locked_documents_rels" USING btree ("plans_id");
  CREATE INDEX "payload_locked_documents_rels_posts_id_idx" ON "cms"."payload_locked_documents_rels" USING btree ("posts_id");
  CREATE INDEX "payload_preferences_key_idx" ON "cms"."payload_preferences" USING btree ("key");
  CREATE INDEX "payload_preferences_updated_at_idx" ON "cms"."payload_preferences" USING btree ("updated_at");
  CREATE INDEX "payload_preferences_created_at_idx" ON "cms"."payload_preferences" USING btree ("created_at");
  CREATE INDEX "payload_preferences_rels_order_idx" ON "cms"."payload_preferences_rels" USING btree ("order");
  CREATE INDEX "payload_preferences_rels_parent_idx" ON "cms"."payload_preferences_rels" USING btree ("parent_id");
  CREATE INDEX "payload_preferences_rels_path_idx" ON "cms"."payload_preferences_rels" USING btree ("path");
  CREATE INDEX "payload_preferences_rels_users_id_idx" ON "cms"."payload_preferences_rels" USING btree ("users_id");
  CREATE INDEX "payload_migrations_updated_at_idx" ON "cms"."payload_migrations" USING btree ("updated_at");
  CREATE INDEX "payload_migrations_created_at_idx" ON "cms"."payload_migrations" USING btree ("created_at");
  CREATE INDEX "landing_blocks_hero_order_idx" ON "cms"."landing_blocks_hero" USING btree ("_order");
  CREATE INDEX "landing_blocks_hero_parent_id_idx" ON "cms"."landing_blocks_hero" USING btree ("_parent_id");
  CREATE INDEX "landing_blocks_hero_path_idx" ON "cms"."landing_blocks_hero" USING btree ("_path");
  CREATE INDEX "landing_blocks_hero_image_idx" ON "cms"."landing_blocks_hero" USING btree ("image_id");
  CREATE INDEX "landing_blocks_stats_items_order_idx" ON "cms"."landing_blocks_stats_items" USING btree ("_order");
  CREATE INDEX "landing_blocks_stats_items_parent_id_idx" ON "cms"."landing_blocks_stats_items" USING btree ("_parent_id");
  CREATE INDEX "landing_blocks_stats_order_idx" ON "cms"."landing_blocks_stats" USING btree ("_order");
  CREATE INDEX "landing_blocks_stats_parent_id_idx" ON "cms"."landing_blocks_stats" USING btree ("_parent_id");
  CREATE INDEX "landing_blocks_stats_path_idx" ON "cms"."landing_blocks_stats" USING btree ("_path");
  CREATE INDEX "landing_blocks_how_it_works_steps_order_idx" ON "cms"."landing_blocks_how_it_works_steps" USING btree ("_order");
  CREATE INDEX "landing_blocks_how_it_works_steps_parent_id_idx" ON "cms"."landing_blocks_how_it_works_steps" USING btree ("_parent_id");
  CREATE INDEX "landing_blocks_how_it_works_order_idx" ON "cms"."landing_blocks_how_it_works" USING btree ("_order");
  CREATE INDEX "landing_blocks_how_it_works_parent_id_idx" ON "cms"."landing_blocks_how_it_works" USING btree ("_parent_id");
  CREATE INDEX "landing_blocks_how_it_works_path_idx" ON "cms"."landing_blocks_how_it_works" USING btree ("_path");
  CREATE INDEX "landing_blocks_features_cards_order_idx" ON "cms"."landing_blocks_features_cards" USING btree ("_order");
  CREATE INDEX "landing_blocks_features_cards_parent_id_idx" ON "cms"."landing_blocks_features_cards" USING btree ("_parent_id");
  CREATE INDEX "landing_blocks_features_cards_image_idx" ON "cms"."landing_blocks_features_cards" USING btree ("image_id");
  CREATE INDEX "landing_blocks_features_order_idx" ON "cms"."landing_blocks_features" USING btree ("_order");
  CREATE INDEX "landing_blocks_features_parent_id_idx" ON "cms"."landing_blocks_features" USING btree ("_parent_id");
  CREATE INDEX "landing_blocks_features_path_idx" ON "cms"."landing_blocks_features" USING btree ("_path");
  CREATE INDEX "landing_blocks_insights_order_idx" ON "cms"."landing_blocks_insights" USING btree ("_order");
  CREATE INDEX "landing_blocks_insights_parent_id_idx" ON "cms"."landing_blocks_insights" USING btree ("_parent_id");
  CREATE INDEX "landing_blocks_insights_path_idx" ON "cms"."landing_blocks_insights" USING btree ("_path");
  CREATE INDEX "landing_blocks_industries_order_idx" ON "cms"."landing_blocks_industries" USING btree ("_order");
  CREATE INDEX "landing_blocks_industries_parent_id_idx" ON "cms"."landing_blocks_industries" USING btree ("_parent_id");
  CREATE INDEX "landing_blocks_industries_path_idx" ON "cms"."landing_blocks_industries" USING btree ("_path");
  CREATE INDEX "landing_blocks_pricing_order_idx" ON "cms"."landing_blocks_pricing" USING btree ("_order");
  CREATE INDEX "landing_blocks_pricing_parent_id_idx" ON "cms"."landing_blocks_pricing" USING btree ("_parent_id");
  CREATE INDEX "landing_blocks_pricing_path_idx" ON "cms"."landing_blocks_pricing" USING btree ("_path");
  CREATE INDEX "landing_blocks_testimonials_items_order_idx" ON "cms"."landing_blocks_testimonials_items" USING btree ("_order");
  CREATE INDEX "landing_blocks_testimonials_items_parent_id_idx" ON "cms"."landing_blocks_testimonials_items" USING btree ("_parent_id");
  CREATE INDEX "landing_blocks_testimonials_items_avatar_idx" ON "cms"."landing_blocks_testimonials_items" USING btree ("avatar_id");
  CREATE INDEX "landing_blocks_testimonials_order_idx" ON "cms"."landing_blocks_testimonials" USING btree ("_order");
  CREATE INDEX "landing_blocks_testimonials_parent_id_idx" ON "cms"."landing_blocks_testimonials" USING btree ("_parent_id");
  CREATE INDEX "landing_blocks_testimonials_path_idx" ON "cms"."landing_blocks_testimonials" USING btree ("_path");
  CREATE INDEX "landing_blocks_faq_items_order_idx" ON "cms"."landing_blocks_faq_items" USING btree ("_order");
  CREATE INDEX "landing_blocks_faq_items_parent_id_idx" ON "cms"."landing_blocks_faq_items" USING btree ("_parent_id");
  CREATE INDEX "landing_blocks_faq_order_idx" ON "cms"."landing_blocks_faq" USING btree ("_order");
  CREATE INDEX "landing_blocks_faq_parent_id_idx" ON "cms"."landing_blocks_faq" USING btree ("_parent_id");
  CREATE INDEX "landing_blocks_faq_path_idx" ON "cms"."landing_blocks_faq" USING btree ("_path");
  CREATE INDEX "landing_blocks_rich_text_order_idx" ON "cms"."landing_blocks_rich_text" USING btree ("_order");
  CREATE INDEX "landing_blocks_rich_text_parent_id_idx" ON "cms"."landing_blocks_rich_text" USING btree ("_parent_id");
  CREATE INDEX "landing_blocks_rich_text_path_idx" ON "cms"."landing_blocks_rich_text" USING btree ("_path");
  CREATE INDEX "landing_blocks_cta_order_idx" ON "cms"."landing_blocks_cta" USING btree ("_order");
  CREATE INDEX "landing_blocks_cta_parent_id_idx" ON "cms"."landing_blocks_cta" USING btree ("_parent_id");
  CREATE INDEX "landing_blocks_cta_path_idx" ON "cms"."landing_blocks_cta" USING btree ("_path");
  CREATE INDEX "landing_blocks_image_order_idx" ON "cms"."landing_blocks_image" USING btree ("_order");
  CREATE INDEX "landing_blocks_image_parent_id_idx" ON "cms"."landing_blocks_image" USING btree ("_parent_id");
  CREATE INDEX "landing_blocks_image_path_idx" ON "cms"."landing_blocks_image" USING btree ("_path");
  CREATE INDEX "landing_blocks_image_image_idx" ON "cms"."landing_blocks_image" USING btree ("image_id");
  CREATE INDEX "landing_rels_order_idx" ON "cms"."landing_rels" USING btree ("order");
  CREATE INDEX "landing_rels_parent_idx" ON "cms"."landing_rels" USING btree ("parent_id");
  CREATE INDEX "landing_rels_path_idx" ON "cms"."landing_rels" USING btree ("path");
  CREATE INDEX "landing_rels_plans_id_idx" ON "cms"."landing_rels" USING btree ("plans_id");
  CREATE INDEX "navbar_links_order_idx" ON "cms"."navbar_links" USING btree ("_order");
  CREATE INDEX "navbar_links_parent_id_idx" ON "cms"."navbar_links" USING btree ("_parent_id");
  CREATE INDEX "navbar_logo_idx" ON "cms"."navbar" USING btree ("logo_id");
  CREATE INDEX "footer_groups_links_order_idx" ON "cms"."footer_groups_links" USING btree ("_order");
  CREATE INDEX "footer_groups_links_parent_id_idx" ON "cms"."footer_groups_links" USING btree ("_parent_id");
  CREATE INDEX "footer_groups_order_idx" ON "cms"."footer_groups" USING btree ("_order");
  CREATE INDEX "footer_groups_parent_id_idx" ON "cms"."footer_groups" USING btree ("_parent_id");
  CREATE INDEX "footer_social_links_order_idx" ON "cms"."footer_social_links" USING btree ("_order");
  CREATE INDEX "footer_social_links_parent_id_idx" ON "cms"."footer_social_links" USING btree ("_parent_id");
  CREATE INDEX "footer_logo_idx" ON "cms"."footer" USING btree ("logo_id");
  CREATE INDEX "pricing_page_faq_items_order_idx" ON "cms"."pricing_page_faq_items" USING btree ("_order");
  CREATE INDEX "pricing_page_faq_items_parent_id_idx" ON "cms"."pricing_page_faq_items" USING btree ("_parent_id");
  CREATE INDEX "pricing_page_rels_order_idx" ON "cms"."pricing_page_rels" USING btree ("order");
  CREATE INDEX "pricing_page_rels_parent_idx" ON "cms"."pricing_page_rels" USING btree ("parent_id");
  CREATE INDEX "pricing_page_rels_path_idx" ON "cms"."pricing_page_rels" USING btree ("path");
  CREATE INDEX "pricing_page_rels_plans_id_idx" ON "cms"."pricing_page_rels" USING btree ("plans_id");`)
}

export async function down({ db, payload, req }: MigrateDownArgs): Promise<void> {
  await db.execute(sql`
   DROP TABLE "cms"."media" CASCADE;
  DROP TABLE "cms"."users_sessions" CASCADE;
  DROP TABLE "cms"."users" CASCADE;
  DROP TABLE "cms"."pages_blocks_hero" CASCADE;
  DROP TABLE "cms"."pages_blocks_stats_items" CASCADE;
  DROP TABLE "cms"."pages_blocks_stats" CASCADE;
  DROP TABLE "cms"."pages_blocks_how_it_works_steps" CASCADE;
  DROP TABLE "cms"."pages_blocks_how_it_works" CASCADE;
  DROP TABLE "cms"."pages_blocks_features_cards" CASCADE;
  DROP TABLE "cms"."pages_blocks_features" CASCADE;
  DROP TABLE "cms"."pages_blocks_insights" CASCADE;
  DROP TABLE "cms"."pages_blocks_industries" CASCADE;
  DROP TABLE "cms"."pages_blocks_pricing" CASCADE;
  DROP TABLE "cms"."pages_blocks_testimonials_items" CASCADE;
  DROP TABLE "cms"."pages_blocks_testimonials" CASCADE;
  DROP TABLE "cms"."pages_blocks_faq_items" CASCADE;
  DROP TABLE "cms"."pages_blocks_faq" CASCADE;
  DROP TABLE "cms"."pages_blocks_rich_text" CASCADE;
  DROP TABLE "cms"."pages_blocks_cta" CASCADE;
  DROP TABLE "cms"."pages_blocks_image" CASCADE;
  DROP TABLE "cms"."pages" CASCADE;
  DROP TABLE "cms"."pages_rels" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_hero" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_stats_items" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_stats" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_how_it_works_steps" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_how_it_works" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_features_cards" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_features" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_insights" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_industries" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_pricing" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_testimonials_items" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_testimonials" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_faq_items" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_faq" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_rich_text" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_cta" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_image" CASCADE;
  DROP TABLE "cms"."_pages_v" CASCADE;
  DROP TABLE "cms"."_pages_v_rels" CASCADE;
  DROP TABLE "cms"."plans_features" CASCADE;
  DROP TABLE "cms"."plans" CASCADE;
  DROP TABLE "cms"."posts" CASCADE;
  DROP TABLE "cms"."_posts_v" CASCADE;
  DROP TABLE "cms"."payload_kv" CASCADE;
  DROP TABLE "cms"."payload_locked_documents" CASCADE;
  DROP TABLE "cms"."payload_locked_documents_rels" CASCADE;
  DROP TABLE "cms"."payload_preferences" CASCADE;
  DROP TABLE "cms"."payload_preferences_rels" CASCADE;
  DROP TABLE "cms"."payload_migrations" CASCADE;
  DROP TABLE "cms"."landing_blocks_hero" CASCADE;
  DROP TABLE "cms"."landing_blocks_stats_items" CASCADE;
  DROP TABLE "cms"."landing_blocks_stats" CASCADE;
  DROP TABLE "cms"."landing_blocks_how_it_works_steps" CASCADE;
  DROP TABLE "cms"."landing_blocks_how_it_works" CASCADE;
  DROP TABLE "cms"."landing_blocks_features_cards" CASCADE;
  DROP TABLE "cms"."landing_blocks_features" CASCADE;
  DROP TABLE "cms"."landing_blocks_insights" CASCADE;
  DROP TABLE "cms"."landing_blocks_industries" CASCADE;
  DROP TABLE "cms"."landing_blocks_pricing" CASCADE;
  DROP TABLE "cms"."landing_blocks_testimonials_items" CASCADE;
  DROP TABLE "cms"."landing_blocks_testimonials" CASCADE;
  DROP TABLE "cms"."landing_blocks_faq_items" CASCADE;
  DROP TABLE "cms"."landing_blocks_faq" CASCADE;
  DROP TABLE "cms"."landing_blocks_rich_text" CASCADE;
  DROP TABLE "cms"."landing_blocks_cta" CASCADE;
  DROP TABLE "cms"."landing_blocks_image" CASCADE;
  DROP TABLE "cms"."landing" CASCADE;
  DROP TABLE "cms"."landing_rels" CASCADE;
  DROP TABLE "cms"."navbar_links" CASCADE;
  DROP TABLE "cms"."navbar" CASCADE;
  DROP TABLE "cms"."footer_groups_links" CASCADE;
  DROP TABLE "cms"."footer_groups" CASCADE;
  DROP TABLE "cms"."footer_social_links" CASCADE;
  DROP TABLE "cms"."footer" CASCADE;
  DROP TABLE "cms"."pricing_page_faq_items" CASCADE;
  DROP TABLE "cms"."pricing_page" CASCADE;
  DROP TABLE "cms"."pricing_page_rels" CASCADE;
  DROP TABLE "cms"."theme" CASCADE;
  DROP TYPE "cms"."enum_pages_blocks_cta_variant";
  DROP TYPE "cms"."enum_pages_blocks_image_size";
  DROP TYPE "cms"."enum_pages_status";
  DROP TYPE "cms"."enum__pages_v_blocks_cta_variant";
  DROP TYPE "cms"."enum__pages_v_blocks_image_size";
  DROP TYPE "cms"."enum__pages_v_version_status";
  DROP TYPE "cms"."enum_posts_category";
  DROP TYPE "cms"."enum_posts_status";
  DROP TYPE "cms"."enum__posts_v_version_category";
  DROP TYPE "cms"."enum__posts_v_version_status";
  DROP TYPE "cms"."enum_landing_blocks_cta_variant";
  DROP TYPE "cms"."enum_landing_blocks_image_size";`)
}
