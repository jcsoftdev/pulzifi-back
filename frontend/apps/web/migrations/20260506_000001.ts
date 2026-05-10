import { MigrateUpArgs, MigrateDownArgs, sql } from '@payloadcms/db-postgres'

export async function up({ db, payload, req }: MigrateUpArgs): Promise<void> {
  await db.execute(sql`
      CREATE TABLE IF NOT EXISTS "cms"."plans_features" (
            "_order" integer NOT NULL,
                  "_parent_id" integer NOT NULL,
                        "id" varchar PRIMARY KEY NOT NULL,
                              "text" varchar,
                                    "included" boolean DEFAULT true
                                        );
                                            CREATE INDEX IF NOT EXISTS "plans_features_order_idx" ON "cms"."plans_features" USING btree ("_order");
                                                CREATE INDEX IF NOT EXISTS "plans_features_parent_id_idx" ON "cms"."plans_features" USING btree ("_parent_id");

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
                                                                                                                      "updated_at" timestamp(3) with time zone DEFAULT now() NOT NULL,
                                                                                                                            "created_at" timestamp(3) with time zone DEFAULT now() NOT NULL
                                                                                                                                );
                                                                                                                                    CREATE INDEX IF NOT EXISTS "plans_created_at_idx" ON "cms"."plans" USING btree ("created_at" DESC NULLS LAST);
                                                                                                                                        CREATE INDEX IF NOT EXISTS "plans_updated_at_idx" ON "cms"."plans" USING btree ("updated_at" DESC NULLS LAST);
                                                                                                                                        
                                                                                                                                            ALTER TABLE "cms"."plans_features" ADD CONSTRAINT "plans_features_parent_id_fk"
                                                                                                                                                  FOREIGN KEY ("_parent_id") REFERENCES "cms"."plans"("id") ON DELETE cascade ON UPDATE no action;
                                                                                                                                                    `)
                                                                                                                                                    }
                                                                                                                                                    
                                                                                                                                                    export async function down({ db, payload, req }: MigrateDownArgs): Promise<void> {
                                                                                                                                                      await db.execute(sql`
                                                                                                                                                          ALTER TABLE "cms"."plans_features" DROP CONSTRAINT IF EXISTS "plans_features_parent_id_fk";
                                                                                                                                                              DROP TABLE IF EXISTS "cms"."plans_features";
                                                                                                                                                                  DROP TABLE IF EXISTS "cms"."plans";
                                                                                                                                                                    `)
                                                                                                                                                                    }
                                                                                                                                                                    
