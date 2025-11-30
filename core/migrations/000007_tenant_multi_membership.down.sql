-- Remove tenant multi-membership tables
DROP TABLE IF EXISTS tenant_chat_messages CASCADE;
DROP TABLE IF EXISTS tenant_announcements CASCADE;
DROP TABLE IF EXISTS tenant_invites CASCADE;
DROP TABLE IF EXISTS tenant_members CASCADE;
