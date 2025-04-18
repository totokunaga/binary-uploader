USE fs_store;

-- Drop tables in reverse order to avoid foreign key constraints
DROP TABLE IF EXISTS `file_chunks`;
DROP TABLE IF EXISTS `files`; 