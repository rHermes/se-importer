package main

// language=SQL
const sqlSetupTables = `IF OBJECT_ID('site', 'U') IS NULL
CREATE TABLE [site]
(
    [id]   SMALLINT IDENTITY (1, 1) PRIMARY KEY,
    [name] NVARCHAR(100) NOT NULL UNIQUE
);

IF OBJECT_ID('user', 'U') IS NULL
CREATE TABLE [user]
(
    [id]              INT            NOT NULL,
    [site_id]         SMALLINT       NOT NULL,
    reputation        INT            NOT NULL,
    creation_date     DATETIMEOFFSET NOT NULL,
    display_name      NVARCHAR(40),
    last_access_date  DATETIMEOFFSET NOT NULL,
    website_url       NVARCHAR(200),
    location          NVARCHAR(100),
    about_me          NVARCHAR(MAX),
    views             INT            NOT NULL,
    up_votes          INT            NOT NULL,
    down_votes        INT            NOT NULL,
    profile_image_url NVARCHAR(200),
    account_id        INT,
    PRIMARY KEY ([id], [site_id]),
    FOREIGN KEY (site_id) REFERENCES [site] ([id]) ON DELETE CASCADE
);
    `

// language=SQL
const sqlInsertSite = `DECLARE @uid SMALLINT
SELECT @uid = (SELECT id FROM [site] WHERE name = @name)

IF @uid IS NULL
    INSERT INTO [site]([name]) OUTPUT inserted.[id] VALUES (@name);
ELSE
    SELECT @uid as [id];`

// language=SQL
const sqlDeleteSite = `DELETE FROM [site] WHERE [name] = @name`