// biome-ignore-all lint/correctness/noUnusedFunctionParameters: generated migration file — payload/req required by interface
import { type MigrateDownArgs, type MigrateUpArgs, sql } from '@payloadcms/db-postgres'

export async function up({ db, payload, req }: MigrateUpArgs): Promise<void> {
  await db.execute(sql`
   CREATE TYPE "cms"."enum_block_library_blocks_hero_dashboard_alerts_tone" AS ENUM('signal', 'amber', 'teal', 'ink');
  CREATE TYPE "cms"."enum_block_library_blocks_hero_kpis_delta_direction" AS ENUM('up', 'down');
  CREATE TYPE "cms"."enum_block_library_blocks_how_it_works_steps_mock_type" AS ENUM('url', 'insight', 'alerts');
  CREATE TYPE "cms"."enum_block_library_blocks_comparison_rows_cells_state" AS ENUM('yes', 'no', 'partial');
  CREATE TYPE "cms"."enum_block_library_blocks_cta_variant" AS ENUM('primary', 'secondary');
  CREATE TYPE "cms"."enum_block_library_blocks_image_size" AS ENUM('full', 'contained');
  CREATE TYPE "cms"."enum_pages_blocks_hero_dashboard_alerts_tone" AS ENUM('signal', 'amber', 'teal', 'ink');
  CREATE TYPE "cms"."enum_pages_blocks_hero_kpis_delta_direction" AS ENUM('up', 'down');
  CREATE TYPE "cms"."enum_pages_blocks_how_it_works_steps_mock_type" AS ENUM('url', 'insight', 'alerts');
  CREATE TYPE "cms"."enum_pages_blocks_comparison_rows_cells_state" AS ENUM('yes', 'no', 'partial');
  CREATE TYPE "cms"."enum__pages_v_blocks_hero_dashboard_alerts_tone" AS ENUM('signal', 'amber', 'teal', 'ink');
  CREATE TYPE "cms"."enum__pages_v_blocks_hero_kpis_delta_direction" AS ENUM('up', 'down');
  CREATE TYPE "cms"."enum__pages_v_blocks_how_it_works_steps_mock_type" AS ENUM('url', 'insight', 'alerts');
  CREATE TYPE "cms"."enum__pages_v_blocks_comparison_rows_cells_state" AS ENUM('yes', 'no', 'partial');
  CREATE TYPE "cms"."enum_footer_social_links_platform" AS ENUM('twitter', 'linkedin', 'github', 'youtube', 'other');
  CREATE TABLE "cms"."block_library_blocks_hero_dashboard_alerts" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"tone" "cms"."enum_block_library_blocks_hero_dashboard_alerts_tone" DEFAULT 'signal',
  	"icon" varchar,
  	"site" varchar NOT NULL,
  	"title" varchar NOT NULL,
  	"detail" varchar,
  	"time" varchar
  );
  
  CREATE TABLE "cms"."block_library_blocks_hero_kpis" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"label" varchar NOT NULL,
  	"value" varchar NOT NULL,
  	"delta" varchar,
  	"delta_direction" "cms"."enum_block_library_blocks_hero_kpis_delta_direction" DEFAULT 'up'
  );
  
  CREATE TABLE "cms"."block_library_blocks_hero" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"eyebrow_badge" varchar,
  	"eyebrow_text" varchar,
  	"headline" varchar NOT NULL,
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
  
  CREATE TABLE "cms"."block_library_blocks_logos_items" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"text" varchar NOT NULL
  );
  
  CREATE TABLE "cms"."block_library_blocks_logos" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"label" varchar,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."block_library_blocks_problem_cards" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"metric" varchar NOT NULL,
  	"label" varchar NOT NULL,
  	"description" varchar NOT NULL
  );
  
  CREATE TABLE "cms"."block_library_blocks_problem" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"eyebrow" varchar,
  	"headline" varchar NOT NULL,
  	"headline_highlight" varchar,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."block_library_blocks_stats_items" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"value" varchar NOT NULL,
  	"label" varchar NOT NULL
  );
  
  CREATE TABLE "cms"."block_library_blocks_stats" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."block_library_blocks_how_it_works_steps" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"step" numeric NOT NULL,
  	"icon" varchar,
  	"title" varchar NOT NULL,
  	"description" varchar NOT NULL,
  	"mock_type" "cms"."enum_block_library_blocks_how_it_works_steps_mock_type" DEFAULT 'url' NOT NULL,
  	"mock_text" varchar
  );
  
  CREATE TABLE "cms"."block_library_blocks_how_it_works" (
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
  
  CREATE TABLE "cms"."block_library_blocks_features_bullets" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"title" varchar NOT NULL,
  	"description" varchar NOT NULL
  );
  
  CREATE TABLE "cms"."block_library_blocks_features_demo_actions" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"label" varchar NOT NULL
  );
  
  CREATE TABLE "cms"."block_library_blocks_features_cards" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"title" varchar NOT NULL,
  	"description" varchar NOT NULL,
  	"image_id" integer
  );
  
  CREATE TABLE "cms"."block_library_blocks_features" (
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
  
  CREATE TABLE "cms"."block_library_blocks_ai_intelligence_tabs_items" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"title" varchar NOT NULL,
  	"body" varchar,
  	"image" varchar
  );
  
  CREATE TABLE "cms"."block_library_blocks_ai_intelligence_tabs" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"label" varchar NOT NULL
  );
  
  CREATE TABLE "cms"."block_library_blocks_ai_intelligence" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"eyebrow" varchar,
  	"headline" varchar NOT NULL,
  	"headline_highlight" varchar,
  	"subheadline" varchar,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."block_library_blocks_insights" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."block_library_blocks_industries_items" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"icon" varchar,
  	"title" varchar NOT NULL,
  	"description" varchar NOT NULL,
  	"real_win" varchar
  );
  
  CREATE TABLE "cms"."block_library_blocks_industries" (
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
  
  CREATE TABLE "cms"."block_library_blocks_comparison_columns" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"name" varchar NOT NULL,
  	"is_us" boolean
  );
  
  CREATE TABLE "cms"."block_library_blocks_comparison_rows_cells" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"state" "cms"."enum_block_library_blocks_comparison_rows_cells_state" DEFAULT 'yes',
  	"note" varchar
  );
  
  CREATE TABLE "cms"."block_library_blocks_comparison_rows" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"feature" varchar NOT NULL
  );
  
  CREATE TABLE "cms"."block_library_blocks_comparison" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"eyebrow" varchar,
  	"headline" varchar,
  	"headline_highlight" varchar,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."block_library_blocks_pricing" (
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
  
  CREATE TABLE "cms"."block_library_blocks_testimonials_items" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"quote" varchar NOT NULL,
  	"author" varchar NOT NULL,
  	"role" varchar,
  	"avatar_id" integer
  );
  
  CREATE TABLE "cms"."block_library_blocks_testimonials" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"eyebrow" varchar,
  	"headline" varchar,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."block_library_blocks_faq_items" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"question" varchar NOT NULL,
  	"answer" varchar NOT NULL
  );
  
  CREATE TABLE "cms"."block_library_blocks_faq" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"eyebrow" varchar,
  	"headline" varchar,
  	"subheadline" varchar,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."block_library_blocks_rich_text" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"content" jsonb NOT NULL,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."block_library_blocks_cta" (
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
  	"variant" "cms"."enum_block_library_blocks_cta_variant" DEFAULT 'primary',
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."block_library_blocks_image" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"image_id" integer NOT NULL,
  	"caption" varchar,
  	"size" "cms"."enum_block_library_blocks_image_size" DEFAULT 'contained',
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."block_library_blocks_login_form" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"headline" varchar,
  	"subheadline" varchar,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."block_library_blocks_register_form" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"headline" varchar,
  	"subheadline" varchar,
  	"trial_badge" varchar,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."block_library_blocks_block_ref" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"ref_id" integer NOT NULL,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."block_library" (
  	"id" serial PRIMARY KEY NOT NULL,
  	"name" varchar NOT NULL,
  	"updated_at" timestamp(3) with time zone DEFAULT now() NOT NULL,
  	"created_at" timestamp(3) with time zone DEFAULT now() NOT NULL
  );
  
  CREATE TABLE "cms"."block_library_rels" (
  	"id" serial PRIMARY KEY NOT NULL,
  	"order" integer,
  	"parent_id" integer NOT NULL,
  	"path" varchar NOT NULL,
  	"plans_id" integer
  );
  
  CREATE TABLE "cms"."pages_blocks_hero_dashboard_alerts" (
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
  
  CREATE TABLE "cms"."pages_blocks_hero_kpis" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"label" varchar,
  	"value" varchar,
  	"delta" varchar,
  	"delta_direction" "cms"."enum_pages_blocks_hero_kpis_delta_direction" DEFAULT 'up'
  );
  
  CREATE TABLE "cms"."pages_blocks_logos_items" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"text" varchar
  );
  
  CREATE TABLE "cms"."pages_blocks_logos" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"label" varchar,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."pages_blocks_problem_cards" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"metric" varchar,
  	"label" varchar,
  	"description" varchar
  );
  
  CREATE TABLE "cms"."pages_blocks_problem" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"eyebrow" varchar,
  	"headline" varchar,
  	"headline_highlight" varchar,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."pages_blocks_features_bullets" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"title" varchar,
  	"description" varchar
  );
  
  CREATE TABLE "cms"."pages_blocks_features_demo_actions" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"label" varchar
  );
  
  CREATE TABLE "cms"."pages_blocks_ai_intelligence_tabs_items" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"title" varchar,
  	"body" varchar,
  	"image" varchar
  );
  
  CREATE TABLE "cms"."pages_blocks_ai_intelligence_tabs" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"label" varchar
  );
  
  CREATE TABLE "cms"."pages_blocks_ai_intelligence" (
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
  
  CREATE TABLE "cms"."pages_blocks_industries_items" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"icon" varchar,
  	"title" varchar,
  	"description" varchar,
  	"real_win" varchar
  );
  
  CREATE TABLE "cms"."pages_blocks_comparison_columns" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"name" varchar,
  	"is_us" boolean
  );
  
  CREATE TABLE "cms"."pages_blocks_comparison_rows_cells" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"state" "cms"."enum_pages_blocks_comparison_rows_cells_state" DEFAULT 'yes',
  	"note" varchar
  );
  
  CREATE TABLE "cms"."pages_blocks_comparison_rows" (
  	"_order" integer NOT NULL,
  	"_parent_id" varchar NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"feature" varchar
  );
  
  CREATE TABLE "cms"."pages_blocks_comparison" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"eyebrow" varchar,
  	"headline" varchar,
  	"headline_highlight" varchar,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."pages_blocks_login_form" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"headline" varchar,
  	"subheadline" varchar,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."pages_blocks_register_form" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"headline" varchar,
  	"subheadline" varchar,
  	"trial_badge" varchar,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."pages_blocks_block_ref" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"ref_id" integer,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."_pages_v_blocks_hero_dashboard_alerts" (
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
  
  CREATE TABLE "cms"."_pages_v_blocks_hero_kpis" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"id" serial PRIMARY KEY NOT NULL,
  	"label" varchar,
  	"value" varchar,
  	"delta" varchar,
  	"delta_direction" "cms"."enum__pages_v_blocks_hero_kpis_delta_direction" DEFAULT 'up',
  	"_uuid" varchar
  );
  
  CREATE TABLE "cms"."_pages_v_blocks_logos_items" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"id" serial PRIMARY KEY NOT NULL,
  	"text" varchar,
  	"_uuid" varchar
  );
  
  CREATE TABLE "cms"."_pages_v_blocks_logos" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" serial PRIMARY KEY NOT NULL,
  	"label" varchar,
  	"_uuid" varchar,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."_pages_v_blocks_problem_cards" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"id" serial PRIMARY KEY NOT NULL,
  	"metric" varchar,
  	"label" varchar,
  	"description" varchar,
  	"_uuid" varchar
  );
  
  CREATE TABLE "cms"."_pages_v_blocks_problem" (
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
  
  CREATE TABLE "cms"."_pages_v_blocks_features_bullets" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"id" serial PRIMARY KEY NOT NULL,
  	"title" varchar,
  	"description" varchar,
  	"_uuid" varchar
  );
  
  CREATE TABLE "cms"."_pages_v_blocks_features_demo_actions" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"id" serial PRIMARY KEY NOT NULL,
  	"label" varchar,
  	"_uuid" varchar
  );
  
  CREATE TABLE "cms"."_pages_v_blocks_ai_intelligence_tabs_items" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"id" serial PRIMARY KEY NOT NULL,
  	"title" varchar,
  	"body" varchar,
  	"image" varchar,
  	"_uuid" varchar
  );
  
  CREATE TABLE "cms"."_pages_v_blocks_ai_intelligence_tabs" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"id" serial PRIMARY KEY NOT NULL,
  	"label" varchar,
  	"_uuid" varchar
  );
  
  CREATE TABLE "cms"."_pages_v_blocks_ai_intelligence" (
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
  
  CREATE TABLE "cms"."_pages_v_blocks_industries_items" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"id" serial PRIMARY KEY NOT NULL,
  	"icon" varchar,
  	"title" varchar,
  	"description" varchar,
  	"real_win" varchar,
  	"_uuid" varchar
  );
  
  CREATE TABLE "cms"."_pages_v_blocks_comparison_columns" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"id" serial PRIMARY KEY NOT NULL,
  	"name" varchar,
  	"is_us" boolean,
  	"_uuid" varchar
  );
  
  CREATE TABLE "cms"."_pages_v_blocks_comparison_rows_cells" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"id" serial PRIMARY KEY NOT NULL,
  	"state" "cms"."enum__pages_v_blocks_comparison_rows_cells_state" DEFAULT 'yes',
  	"note" varchar,
  	"_uuid" varchar
  );
  
  CREATE TABLE "cms"."_pages_v_blocks_comparison_rows" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"id" serial PRIMARY KEY NOT NULL,
  	"feature" varchar,
  	"_uuid" varchar
  );
  
  CREATE TABLE "cms"."_pages_v_blocks_comparison" (
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
  
  CREATE TABLE "cms"."_pages_v_blocks_login_form" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" serial PRIMARY KEY NOT NULL,
  	"headline" varchar,
  	"subheadline" varchar,
  	"_uuid" varchar,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."_pages_v_blocks_register_form" (
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
  
  CREATE TABLE "cms"."_pages_v_blocks_block_ref" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"_path" text NOT NULL,
  	"id" serial PRIMARY KEY NOT NULL,
  	"ref_id" integer,
  	"_uuid" varchar,
  	"block_name" varchar
  );
  
  CREATE TABLE "cms"."login_page_bullets" (
  	"_order" integer NOT NULL,
  	"_parent_id" integer NOT NULL,
  	"id" varchar PRIMARY KEY NOT NULL,
  	"text" varchar NOT NULL
  );
  
  CREATE TABLE "cms"."login_page" (
  	"id" serial PRIMARY KEY NOT NULL,
  	"headline" varchar DEFAULT 'Know before your competitors move.',
  	"subheadline" varchar DEFAULT 'Join 2,500+ teams who react first.',
  	"welcome_headline" varchar DEFAULT 'Welcome back',
  	"welcome_subline" varchar DEFAULT 'Sign in to your account',
  	"updated_at" timestamp(3) with time zone,
  	"created_at" timestamp(3) with time zone
  );
  
  CREATE TABLE "cms"."contact_page" (
  	"id" serial PRIMARY KEY NOT NULL,
  	"headline" varchar DEFAULT 'Get in touch',
  	"subheadline" varchar DEFAULT 'We''re here to help. Our team responds within one business day.',
  	"email" varchar DEFAULT 'support@pulzifi.com',
  	"address" varchar DEFAULT 'Boise, ID',
  	"updated_at" timestamp(3) with time zone,
  	"created_at" timestamp(3) with time zone
  );

  CREATE TABLE IF NOT EXISTS "cms"."payload_migrations" (
  	"id" serial PRIMARY KEY NOT NULL,
  	"name" varchar,
  	"batch" numeric,
  	"updated_at" timestamp(3) with time zone NOT NULL DEFAULT now(),
  	"created_at" timestamp(3) with time zone NOT NULL DEFAULT now()
  );
`)
}

export async function down({ db, payload, req }: MigrateDownArgs): Promise<void> {
  await db.execute(sql`
   CREATE TYPE "cms"."enum_landing_blocks_cta_variant" AS ENUM('primary', 'secondary');
  CREATE TYPE "cms"."enum_landing_blocks_image_size" AS ENUM('full', 'contained');
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
  
  ALTER TABLE "cms"."block_library_blocks_hero_dashboard_alerts" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_hero_kpis" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_hero" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_logos_items" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_logos" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_problem_cards" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_problem" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_stats_items" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_stats" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_how_it_works_steps" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_how_it_works" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_features_bullets" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_features_demo_actions" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_features_cards" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_features" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_ai_intelligence_tabs_items" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_ai_intelligence_tabs" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_ai_intelligence" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_insights" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_industries_items" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_industries" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_comparison_columns" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_comparison_rows_cells" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_comparison_rows" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_comparison" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_pricing" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_testimonials_items" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_testimonials" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_faq_items" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_faq" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_rich_text" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_cta" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_image" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_login_form" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_register_form" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_blocks_block_ref" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."block_library_rels" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."pages_blocks_hero_dashboard_alerts" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."pages_blocks_hero_kpis" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."pages_blocks_logos_items" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."pages_blocks_logos" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."pages_blocks_problem_cards" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."pages_blocks_problem" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."pages_blocks_features_bullets" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."pages_blocks_features_demo_actions" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."pages_blocks_ai_intelligence_tabs_items" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."pages_blocks_ai_intelligence_tabs" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."pages_blocks_ai_intelligence" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."pages_blocks_industries_items" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."pages_blocks_comparison_columns" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."pages_blocks_comparison_rows_cells" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."pages_blocks_comparison_rows" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."pages_blocks_comparison" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."pages_blocks_login_form" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."pages_blocks_register_form" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."pages_blocks_block_ref" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."_pages_v_blocks_hero_dashboard_alerts" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."_pages_v_blocks_hero_kpis" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."_pages_v_blocks_logos_items" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."_pages_v_blocks_logos" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."_pages_v_blocks_problem_cards" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."_pages_v_blocks_problem" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."_pages_v_blocks_features_bullets" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."_pages_v_blocks_features_demo_actions" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."_pages_v_blocks_ai_intelligence_tabs_items" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."_pages_v_blocks_ai_intelligence_tabs" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."_pages_v_blocks_ai_intelligence" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."_pages_v_blocks_industries_items" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."_pages_v_blocks_comparison_columns" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."_pages_v_blocks_comparison_rows_cells" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."_pages_v_blocks_comparison_rows" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."_pages_v_blocks_comparison" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."_pages_v_blocks_login_form" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."_pages_v_blocks_register_form" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."_pages_v_blocks_block_ref" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."login_page_bullets" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."login_page" DISABLE ROW LEVEL SECURITY;
  ALTER TABLE "cms"."contact_page" DISABLE ROW LEVEL SECURITY;
  DROP TABLE "cms"."block_library_blocks_hero_dashboard_alerts" CASCADE;
  DROP TABLE "cms"."block_library_blocks_hero_kpis" CASCADE;
  DROP TABLE "cms"."block_library_blocks_hero" CASCADE;
  DROP TABLE "cms"."block_library_blocks_logos_items" CASCADE;
  DROP TABLE "cms"."block_library_blocks_logos" CASCADE;
  DROP TABLE "cms"."block_library_blocks_problem_cards" CASCADE;
  DROP TABLE "cms"."block_library_blocks_problem" CASCADE;
  DROP TABLE "cms"."block_library_blocks_stats_items" CASCADE;
  DROP TABLE "cms"."block_library_blocks_stats" CASCADE;
  DROP TABLE "cms"."block_library_blocks_how_it_works_steps" CASCADE;
  DROP TABLE "cms"."block_library_blocks_how_it_works" CASCADE;
  DROP TABLE "cms"."block_library_blocks_features_bullets" CASCADE;
  DROP TABLE "cms"."block_library_blocks_features_demo_actions" CASCADE;
  DROP TABLE "cms"."block_library_blocks_features_cards" CASCADE;
  DROP TABLE "cms"."block_library_blocks_features" CASCADE;
  DROP TABLE "cms"."block_library_blocks_ai_intelligence_tabs_items" CASCADE;
  DROP TABLE "cms"."block_library_blocks_ai_intelligence_tabs" CASCADE;
  DROP TABLE "cms"."block_library_blocks_ai_intelligence" CASCADE;
  DROP TABLE "cms"."block_library_blocks_insights" CASCADE;
  DROP TABLE "cms"."block_library_blocks_industries_items" CASCADE;
  DROP TABLE "cms"."block_library_blocks_industries" CASCADE;
  DROP TABLE "cms"."block_library_blocks_comparison_columns" CASCADE;
  DROP TABLE "cms"."block_library_blocks_comparison_rows_cells" CASCADE;
  DROP TABLE "cms"."block_library_blocks_comparison_rows" CASCADE;
  DROP TABLE "cms"."block_library_blocks_comparison" CASCADE;
  DROP TABLE "cms"."block_library_blocks_pricing" CASCADE;
  DROP TABLE "cms"."block_library_blocks_testimonials_items" CASCADE;
  DROP TABLE "cms"."block_library_blocks_testimonials" CASCADE;
  DROP TABLE "cms"."block_library_blocks_faq_items" CASCADE;
  DROP TABLE "cms"."block_library_blocks_faq" CASCADE;
  DROP TABLE "cms"."block_library_blocks_rich_text" CASCADE;
  DROP TABLE "cms"."block_library_blocks_cta" CASCADE;
  DROP TABLE "cms"."block_library_blocks_image" CASCADE;
  DROP TABLE "cms"."block_library_blocks_login_form" CASCADE;
  DROP TABLE "cms"."block_library_blocks_register_form" CASCADE;
  DROP TABLE "cms"."block_library_blocks_block_ref" CASCADE;
  DROP TABLE "cms"."block_library" CASCADE;
  DROP TABLE "cms"."block_library_rels" CASCADE;
  DROP TABLE "cms"."pages_blocks_hero_dashboard_alerts" CASCADE;
  DROP TABLE "cms"."pages_blocks_hero_kpis" CASCADE;
  DROP TABLE "cms"."pages_blocks_logos_items" CASCADE;
  DROP TABLE "cms"."pages_blocks_logos" CASCADE;
  DROP TABLE "cms"."pages_blocks_problem_cards" CASCADE;
  DROP TABLE "cms"."pages_blocks_problem" CASCADE;
  DROP TABLE "cms"."pages_blocks_features_bullets" CASCADE;
  DROP TABLE "cms"."pages_blocks_features_demo_actions" CASCADE;
  DROP TABLE "cms"."pages_blocks_ai_intelligence_tabs_items" CASCADE;
  DROP TABLE "cms"."pages_blocks_ai_intelligence_tabs" CASCADE;
  DROP TABLE "cms"."pages_blocks_ai_intelligence" CASCADE;
  DROP TABLE "cms"."pages_blocks_industries_items" CASCADE;
  DROP TABLE "cms"."pages_blocks_comparison_columns" CASCADE;
  DROP TABLE "cms"."pages_blocks_comparison_rows_cells" CASCADE;
  DROP TABLE "cms"."pages_blocks_comparison_rows" CASCADE;
  DROP TABLE "cms"."pages_blocks_comparison" CASCADE;
  DROP TABLE "cms"."pages_blocks_login_form" CASCADE;
  DROP TABLE "cms"."pages_blocks_register_form" CASCADE;
  DROP TABLE "cms"."pages_blocks_block_ref" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_hero_dashboard_alerts" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_hero_kpis" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_logos_items" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_logos" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_problem_cards" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_problem" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_features_bullets" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_features_demo_actions" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_ai_intelligence_tabs_items" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_ai_intelligence_tabs" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_ai_intelligence" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_industries_items" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_comparison_columns" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_comparison_rows_cells" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_comparison_rows" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_comparison" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_login_form" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_register_form" CASCADE;
  DROP TABLE "cms"."_pages_v_blocks_block_ref" CASCADE;
  DROP TABLE "cms"."login_page_bullets" CASCADE;
  DROP TABLE "cms"."login_page" CASCADE;
  DROP TABLE "cms"."contact_page" CASCADE;
  ALTER TABLE "cms"."payload_locked_documents_rels" DROP CONSTRAINT "payload_locked_documents_rels_block_library_fk";
  
  DROP INDEX "cms"."payload_locked_documents_rels_block_library_id_idx";
  ALTER TABLE "cms"."plans_features" ALTER COLUMN "text" SET NOT NULL;
  ALTER TABLE "cms"."plans" ALTER COLUMN "highlighted" SET DEFAULT false;
  ALTER TABLE "cms"."navbar" ALTER COLUMN "signin_label" DROP DEFAULT;
  ALTER TABLE "cms"."navbar" ALTER COLUMN "signin_href" DROP DEFAULT;
  ALTER TABLE "cms"."navbar" ALTER COLUMN "primary_cta_label" DROP DEFAULT;
  ALTER TABLE "cms"."navbar" ALTER COLUMN "primary_cta_href" DROP DEFAULT;
  ALTER TABLE "cms"."footer_social_links" ALTER COLUMN "platform" SET DATA TYPE varchar;
  ALTER TABLE "cms"."pages_blocks_hero" ADD COLUMN "cta_label" varchar;
  ALTER TABLE "cms"."pages_blocks_hero" ADD COLUMN "cta_href" varchar;
  ALTER TABLE "cms"."_pages_v_blocks_hero" ADD COLUMN "cta_label" varchar;
  ALTER TABLE "cms"."_pages_v_blocks_hero" ADD COLUMN "cta_href" varchar;
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
  ALTER TABLE "cms"."pricing_page_faq_items" ADD CONSTRAINT "pricing_page_faq_items_parent_id_fk" FOREIGN KEY ("_parent_id") REFERENCES "cms"."pricing_page"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."pricing_page_rels" ADD CONSTRAINT "pricing_page_rels_parent_fk" FOREIGN KEY ("parent_id") REFERENCES "cms"."pricing_page"("id") ON DELETE cascade ON UPDATE no action;
  ALTER TABLE "cms"."pricing_page_rels" ADD CONSTRAINT "pricing_page_rels_plans_fk" FOREIGN KEY ("plans_id") REFERENCES "cms"."plans"("id") ON DELETE cascade ON UPDATE no action;
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
  CREATE INDEX "pricing_page_faq_items_order_idx" ON "cms"."pricing_page_faq_items" USING btree ("_order");
  CREATE INDEX "pricing_page_faq_items_parent_id_idx" ON "cms"."pricing_page_faq_items" USING btree ("_parent_id");
  CREATE INDEX "pricing_page_rels_order_idx" ON "cms"."pricing_page_rels" USING btree ("order");
  CREATE INDEX "pricing_page_rels_parent_idx" ON "cms"."pricing_page_rels" USING btree ("parent_id");
  CREATE INDEX "pricing_page_rels_path_idx" ON "cms"."pricing_page_rels" USING btree ("path");
  CREATE INDEX "pricing_page_rels_plans_id_idx" ON "cms"."pricing_page_rels" USING btree ("plans_id");
  ALTER TABLE "cms"."pages_blocks_hero" DROP COLUMN "eyebrow_text";
  ALTER TABLE "cms"."pages_blocks_hero" DROP COLUMN "headline_highlight";
  ALTER TABLE "cms"."pages_blocks_hero" DROP COLUMN "primary_cta_label";
  ALTER TABLE "cms"."pages_blocks_hero" DROP COLUMN "primary_cta_href";
  ALTER TABLE "cms"."pages_blocks_hero" DROP COLUMN "secondary_cta_label";
  ALTER TABLE "cms"."pages_blocks_hero" DROP COLUMN "secondary_cta_href";
  ALTER TABLE "cms"."pages_blocks_hero" DROP COLUMN "trust_line";
  ALTER TABLE "cms"."pages_blocks_hero" DROP COLUMN "ai_insight_title";
  ALTER TABLE "cms"."pages_blocks_hero" DROP COLUMN "ai_insight_body";
  ALTER TABLE "cms"."pages_blocks_how_it_works_steps" DROP COLUMN "icon";
  ALTER TABLE "cms"."pages_blocks_how_it_works_steps" DROP COLUMN "mock_type";
  ALTER TABLE "cms"."pages_blocks_how_it_works_steps" DROP COLUMN "mock_text";
  ALTER TABLE "cms"."pages_blocks_how_it_works" DROP COLUMN "eyebrow";
  ALTER TABLE "cms"."pages_blocks_how_it_works" DROP COLUMN "headline";
  ALTER TABLE "cms"."pages_blocks_how_it_works" DROP COLUMN "headline_highlight";
  ALTER TABLE "cms"."pages_blocks_how_it_works" DROP COLUMN "subheadline";
  ALTER TABLE "cms"."pages_blocks_features" DROP COLUMN "eyebrow";
  ALTER TABLE "cms"."pages_blocks_features" DROP COLUMN "headline";
  ALTER TABLE "cms"."pages_blocks_features" DROP COLUMN "headline_highlight";
  ALTER TABLE "cms"."pages_blocks_features" DROP COLUMN "intro";
  ALTER TABLE "cms"."pages_blocks_features" DROP COLUMN "demo_title";
  ALTER TABLE "cms"."pages_blocks_features" DROP COLUMN "demo_badge";
  ALTER TABLE "cms"."pages_blocks_features" DROP COLUMN "demo_site";
  ALTER TABLE "cms"."pages_blocks_features" DROP COLUMN "demo_change";
  ALTER TABLE "cms"."pages_blocks_features" DROP COLUMN "demo_analysis";
  ALTER TABLE "cms"."pages_blocks_industries" DROP COLUMN "compact_mode";
  ALTER TABLE "cms"."pages_blocks_industries" DROP COLUMN "eyebrow";
  ALTER TABLE "cms"."pages_blocks_industries" DROP COLUMN "headline";
  ALTER TABLE "cms"."pages_blocks_industries" DROP COLUMN "headline_highlight";
  ALTER TABLE "cms"."pages_blocks_industries" DROP COLUMN "subheadline";
  ALTER TABLE "cms"."pages_blocks_pricing" DROP COLUMN "billing_monthly_label";
  ALTER TABLE "cms"."pages_blocks_pricing" DROP COLUMN "billing_annual_label";
  ALTER TABLE "cms"."pages_blocks_pricing" DROP COLUMN "billing_annual_badge";
  ALTER TABLE "cms"."pages_blocks_pricing" DROP COLUMN "billing_annual_note";
  ALTER TABLE "cms"."pages_blocks_pricing" DROP COLUMN "compare_plans_headline";
  ALTER TABLE "cms"."pages_blocks_pricing" DROP COLUMN "features_label";
  ALTER TABLE "cms"."pages_blocks_testimonials" DROP COLUMN "eyebrow";
  ALTER TABLE "cms"."pages_blocks_testimonials" DROP COLUMN "headline";
  ALTER TABLE "cms"."pages_blocks_faq" DROP COLUMN "eyebrow";
  ALTER TABLE "cms"."pages_blocks_faq" DROP COLUMN "headline";
  ALTER TABLE "cms"."pages_blocks_faq" DROP COLUMN "subheadline";
  ALTER TABLE "cms"."_pages_v_blocks_hero" DROP COLUMN "eyebrow_badge";
  ALTER TABLE "cms"."_pages_v_blocks_hero" DROP COLUMN "eyebrow_text";
  ALTER TABLE "cms"."_pages_v_blocks_hero" DROP COLUMN "headline_highlight";
  ALTER TABLE "cms"."_pages_v_blocks_hero" DROP COLUMN "primary_cta_label";
  ALTER TABLE "cms"."_pages_v_blocks_hero" DROP COLUMN "primary_cta_href";
  ALTER TABLE "cms"."_pages_v_blocks_hero" DROP COLUMN "secondary_cta_label";
  ALTER TABLE "cms"."_pages_v_blocks_hero" DROP COLUMN "secondary_cta_href";
  ALTER TABLE "cms"."_pages_v_blocks_hero" DROP COLUMN "trust_line";
  ALTER TABLE "cms"."_pages_v_blocks_hero" DROP COLUMN "ai_insight_title";
  ALTER TABLE "cms"."_pages_v_blocks_hero" DROP COLUMN "ai_insight_body";
  ALTER TABLE "cms"."_pages_v_blocks_how_it_works_steps" DROP COLUMN "icon";
  ALTER TABLE "cms"."_pages_v_blocks_how_it_works_steps" DROP COLUMN "mock_type";
  ALTER TABLE "cms"."_pages_v_blocks_how_it_works_steps" DROP COLUMN "mock_text";
  ALTER TABLE "cms"."_pages_v_blocks_how_it_works" DROP COLUMN "eyebrow";
  ALTER TABLE "cms"."_pages_v_blocks_how_it_works" DROP COLUMN "headline";
  ALTER TABLE "cms"."_pages_v_blocks_how_it_works" DROP COLUMN "headline_highlight";
  ALTER TABLE "cms"."_pages_v_blocks_how_it_works" DROP COLUMN "subheadline";
  ALTER TABLE "cms"."_pages_v_blocks_features" DROP COLUMN "eyebrow";
  ALTER TABLE "cms"."_pages_v_blocks_features" DROP COLUMN "headline";
  ALTER TABLE "cms"."_pages_v_blocks_features" DROP COLUMN "headline_highlight";
  ALTER TABLE "cms"."_pages_v_blocks_features" DROP COLUMN "intro";
  ALTER TABLE "cms"."_pages_v_blocks_features" DROP COLUMN "demo_title";
  ALTER TABLE "cms"."_pages_v_blocks_features" DROP COLUMN "demo_badge";
  ALTER TABLE "cms"."_pages_v_blocks_features" DROP COLUMN "demo_site";
  ALTER TABLE "cms"."_pages_v_blocks_features" DROP COLUMN "demo_change";
  ALTER TABLE "cms"."_pages_v_blocks_features" DROP COLUMN "demo_analysis";
  ALTER TABLE "cms"."_pages_v_blocks_industries" DROP COLUMN "compact_mode";
  ALTER TABLE "cms"."_pages_v_blocks_industries" DROP COLUMN "eyebrow";
  ALTER TABLE "cms"."_pages_v_blocks_industries" DROP COLUMN "headline";
  ALTER TABLE "cms"."_pages_v_blocks_industries" DROP COLUMN "headline_highlight";
  ALTER TABLE "cms"."_pages_v_blocks_industries" DROP COLUMN "subheadline";
  ALTER TABLE "cms"."_pages_v_blocks_pricing" DROP COLUMN "billing_monthly_label";
  ALTER TABLE "cms"."_pages_v_blocks_pricing" DROP COLUMN "billing_annual_label";
  ALTER TABLE "cms"."_pages_v_blocks_pricing" DROP COLUMN "billing_annual_badge";
  ALTER TABLE "cms"."_pages_v_blocks_pricing" DROP COLUMN "billing_annual_note";
  ALTER TABLE "cms"."_pages_v_blocks_pricing" DROP COLUMN "compare_plans_headline";
  ALTER TABLE "cms"."_pages_v_blocks_pricing" DROP COLUMN "features_label";
  ALTER TABLE "cms"."_pages_v_blocks_testimonials" DROP COLUMN "eyebrow";
  ALTER TABLE "cms"."_pages_v_blocks_testimonials" DROP COLUMN "headline";
  ALTER TABLE "cms"."_pages_v_blocks_faq" DROP COLUMN "eyebrow";
  ALTER TABLE "cms"."_pages_v_blocks_faq" DROP COLUMN "headline";
  ALTER TABLE "cms"."_pages_v_blocks_faq" DROP COLUMN "subheadline";
  ALTER TABLE "cms"."plans" DROP COLUMN "price_annual";
  ALTER TABLE "cms"."payload_locked_documents_rels" DROP COLUMN "block_library_id";
  DROP TYPE "cms"."enum_block_library_blocks_hero_dashboard_alerts_tone";
  DROP TYPE "cms"."enum_block_library_blocks_hero_kpis_delta_direction";
  DROP TYPE "cms"."enum_block_library_blocks_how_it_works_steps_mock_type";
  DROP TYPE "cms"."enum_block_library_blocks_comparison_rows_cells_state";
  DROP TYPE "cms"."enum_block_library_blocks_cta_variant";
  DROP TYPE "cms"."enum_block_library_blocks_image_size";
  DROP TYPE "cms"."enum_pages_blocks_hero_dashboard_alerts_tone";
  DROP TYPE "cms"."enum_pages_blocks_hero_kpis_delta_direction";
  DROP TYPE "cms"."enum_pages_blocks_how_it_works_steps_mock_type";
  DROP TYPE "cms"."enum_pages_blocks_comparison_rows_cells_state";
  DROP TYPE "cms"."enum__pages_v_blocks_hero_dashboard_alerts_tone";
  DROP TYPE "cms"."enum__pages_v_blocks_hero_kpis_delta_direction";
  DROP TYPE "cms"."enum__pages_v_blocks_how_it_works_steps_mock_type";
  DROP TYPE "cms"."enum__pages_v_blocks_comparison_rows_cells_state";
  DROP TYPE "cms"."enum_footer_social_links_platform";`)
}
