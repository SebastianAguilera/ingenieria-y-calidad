-- Enums
CREATE TYPE "estimate_source" AS ENUM ('direct', 'planning_poker');
CREATE TYPE "project_status" AS ENUM ('planning', 'active', 'on_hold', 'completed', 'cancelled');
CREATE TYPE "member_role" AS ENUM ('product_architect', 'agile_enabler', 'product_builder');
CREATE TYPE "priority_level" AS ENUM ('low', 'medium', 'high', 'critical');
CREATE TYPE "backlog_item_status" AS ENUM ('backlog', 'ready', 'in_sprint', 'in_progress', 'done', 'cancelled');
CREATE TYPE "sprint_status" AS ENUM ('planned', 'active', 'closed');
CREATE TYPE "task_status" AS ENUM ('todo', 'in_progress', 'done');
CREATE TYPE "poker_session_status" AS ENUM ('voting', 'revealed', 'closed');
CREATE TYPE "severity_level" AS ENUM ('minor', 'major', 'critical', 'blocker');
CREATE TYPE "defect_status" AS ENUM ('open', 'in_progress', 'resolved', 'closed', 'rejected');

-- Tables
CREATE TABLE "users" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "full_name" varchar(150) NOT NULL,
  "email" varchar(255) UNIQUE NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "project" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "name" varchar(150) NOT NULL,
  "description" text,
  "status" project_status NOT NULL DEFAULT 'planning',
  "start_date" date,
  "end_date" date,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now()),
  CONSTRAINT "chk_project_dates" CHECK (end_date IS NULL OR end_date >= start_date)
);

CREATE TABLE "project_member" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "project_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "role" member_role NOT NULL DEFAULT 'product_builder',
  "joined_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "backlog_item" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "project_id" uuid NOT NULL,
  "title" varchar(200) NOT NULL,
  "description" text,
  "priority" priority_level NOT NULL DEFAULT 'medium',
  "status" backlog_item_status NOT NULL DEFAULT 'backlog',
  "story_points" integer,
  "estimated_hours" numeric(6,2),
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now()),
  CONSTRAINT "chk_backlog_item_sp" CHECK (story_points IS NULL OR story_points >= 0),
  CONSTRAINT "chk_backlog_item_hours" CHECK (estimated_hours IS NULL OR estimated_hours >= 0)
);

CREATE TABLE "acceptance_criterion" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "backlog_item_id" uuid NOT NULL,
  "description" text NOT NULL,
  "position" integer NOT NULL DEFAULT 1,
  "is_met" boolean NOT NULL DEFAULT false
);

CREATE TABLE "task" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "backlog_item_id" uuid NOT NULL,
  "assignee_id" uuid,
  "title" varchar(200) NOT NULL,
  "description" text,
  "status" task_status NOT NULL DEFAULT 'todo',
  "estimated_hours" numeric(6,2),
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  CONSTRAINT "chk_task_hours" CHECK (estimated_hours IS NULL OR estimated_hours >= 0)
);

CREATE TABLE "sprint" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "project_id" uuid NOT NULL,
  "number" integer NOT NULL,
  "name" varchar(100),
  "goal" text,
  "status" sprint_status NOT NULL DEFAULT 'planned',
  "start_date" date NOT NULL,
  "end_date" date NOT NULL,
  "closed_at" timestamptz,
  CONSTRAINT "chk_sprint_dates" CHECK (end_date >= start_date)
);

CREATE TABLE "sprint_item" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "sprint_id" uuid NOT NULL,
  "backlog_item_id" uuid NOT NULL,
  "planned_story_points" integer,
  "assigned_at" timestamptz NOT NULL DEFAULT (now()),
  "completed_at" timestamptz
);

CREATE TABLE "poker_session" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "backlog_item_id" uuid NOT NULL,
  "created_by" uuid NOT NULL,
  "status" poker_session_status NOT NULL DEFAULT 'voting',
  "current_round" integer NOT NULL DEFAULT 1,
  "agreed_points" integer,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "closed_at" timestamptz
);

CREATE TABLE "poker_round" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "session_id" uuid NOT NULL,
  "round_number" integer NOT NULL,
  "revealed_at" timestamptz
);

CREATE TABLE "poker_vote" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "round_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "points" integer NOT NULL,
  "voted_at" timestamptz NOT NULL DEFAULT (now()),
  CONSTRAINT "chk_poker_vote_points" CHECK (points >= 0)
);

CREATE TABLE "effort_log" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "backlog_item_id" uuid NOT NULL,
  "task_id" uuid,
  "sprint_id" uuid,
  "user_id" uuid NOT NULL,
  "log_date" date NOT NULL,
  "activity" text NOT NULL,
  "hours" numeric(5,2) NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  CONSTRAINT "chk_effort_log_hours" CHECK (hours > 0 AND hours <= 24)
);

CREATE TABLE "defect" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "backlog_item_id" uuid NOT NULL,
  "detected_sprint_id" uuid NOT NULL,
  "resolved_sprint_id" uuid,
  "reported_by" uuid,
  "description" text NOT NULL,
  "severity" severity_level NOT NULL DEFAULT 'minor',
  "status" defect_status NOT NULL DEFAULT 'open',
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "resolved_at" timestamptz
);

CREATE TABLE "sprint_metrics_snapshot" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "sprint_id" uuid UNIQUE NOT NULL,
  "planned_story_points" integer NOT NULL,
  "completed_story_points" integer NOT NULL,
  "velocity" numeric(6,2) NOT NULL,
  "stories_planned" integer NOT NULL,
  "stories_completed" integer NOT NULL,
  "completion_percentage" numeric(5,2) NOT NULL,
  "estimated_hours" numeric(8,2) NOT NULL,
  "actual_hours" numeric(8,2) NOT NULL,
  "effort_deviation_percentage" numeric(7,2),
  "defects_detected" integer NOT NULL DEFAULT 0,
  "defects_resolved" integer NOT NULL DEFAULT 0,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "story_estimate" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "backlog_item_id" uuid NOT NULL,
  "estimated_by" uuid NOT NULL,
  "points" integer NOT NULL,
  "source" estimate_source NOT NULL DEFAULT 'direct',
  "estimated_at" timestamptz NOT NULL DEFAULT (now()),
  CONSTRAINT "chk_story_estimate_fibonacci" CHECK (points IN (1, 2, 3, 5, 8, 13, 21))
);

-- Indexes
CREATE UNIQUE INDEX ON "project_member" ("project_id", "user_id");
CREATE INDEX ON "backlog_item" ("project_id");
CREATE INDEX ON "backlog_item" ("project_id", "status");
CREATE UNIQUE INDEX ON "acceptance_criterion" ("backlog_item_id", "position");
CREATE UNIQUE INDEX ON "sprint" ("project_id", "number");
CREATE UNIQUE INDEX ON "sprint_item" ("sprint_id", "backlog_item_id");
CREATE UNIQUE INDEX ON "poker_round" ("session_id", "round_number");
CREATE UNIQUE INDEX ON "poker_vote" ("round_id", "user_id");
CREATE INDEX ON "effort_log" ("backlog_item_id");
CREATE INDEX ON "effort_log" ("sprint_id", "user_id");
CREATE INDEX ON "defect" ("backlog_item_id");
CREATE INDEX ON "defect" ("detected_sprint_id");
CREATE INDEX ON "story_estimate" ("backlog_item_id", "estimated_at");

-- Foreign Keys
ALTER TABLE "project_member" ADD FOREIGN KEY ("project_id") REFERENCES "project" ("id") ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE "project_member" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE "backlog_item" ADD FOREIGN KEY ("project_id") REFERENCES "project" ("id") ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE "acceptance_criterion" ADD FOREIGN KEY ("backlog_item_id") REFERENCES "backlog_item" ("id") ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE "task" ADD FOREIGN KEY ("backlog_item_id") REFERENCES "backlog_item" ("id") ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE "task" ADD FOREIGN KEY ("assignee_id") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE CASCADE;
ALTER TABLE "sprint" ADD FOREIGN KEY ("project_id") REFERENCES "project" ("id") ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE "sprint_item" ADD FOREIGN KEY ("sprint_id") REFERENCES "sprint" ("id") ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE "sprint_item" ADD FOREIGN KEY ("backlog_item_id") REFERENCES "backlog_item" ("id") ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE "poker_session" ADD FOREIGN KEY ("backlog_item_id") REFERENCES "backlog_item" ("id") ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE "poker_session" ADD FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE "poker_round" ADD FOREIGN KEY ("session_id") REFERENCES "poker_session" ("id") ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE "poker_vote" ADD FOREIGN KEY ("round_id") REFERENCES "poker_round" ("id") ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE "poker_vote" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE "effort_log" ADD FOREIGN KEY ("backlog_item_id") REFERENCES "backlog_item" ("id") ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE "effort_log" ADD FOREIGN KEY ("task_id") REFERENCES "task" ("id") ON DELETE SET NULL ON UPDATE CASCADE;
ALTER TABLE "effort_log" ADD FOREIGN KEY ("sprint_id") REFERENCES "sprint" ("id") ON DELETE SET NULL ON UPDATE CASCADE;
ALTER TABLE "effort_log" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE "defect" ADD FOREIGN KEY ("backlog_item_id") REFERENCES "backlog_item" ("id") ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE "defect" ADD FOREIGN KEY ("detected_sprint_id") REFERENCES "sprint" ("id") ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE "defect" ADD FOREIGN KEY ("resolved_sprint_id") REFERENCES "sprint" ("id") ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE "defect" ADD FOREIGN KEY ("reported_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE CASCADE;
ALTER TABLE "sprint_metrics_snapshot" ADD FOREIGN KEY ("sprint_id") REFERENCES "sprint" ("id") ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE "story_estimate" ADD FOREIGN KEY ("backlog_item_id") REFERENCES "backlog_item" ("id") ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE "story_estimate" ADD FOREIGN KEY ("estimated_by") REFERENCES "users" ("id") ON DELETE RESTRICT ON UPDATE CASCADE;