/* Remove migration-related tables that are only updated through the server to track which
   migrations have been applied */
DROP TABLE IF EXISTS db_lock;
DROP TABLE IF EXISTS db_migrations;

/* The security update check in the server may update the LastSecurityTime system value. To
   avoid any spurious difference in the migrations, we update it to a fixed value. */
UPDATE Systems SET Value = 1 WHERE Name = 'LastSecurityTime';

/* The server migration may contain a row in the Systems table marking the onboarding as complete.
   There are no migrations related to this, so we can simply drop it here. */
DELETE FROM Systems WHERE Name = 'FirstAdminSetupComplete';

/* The server migration contains an in-app migration that add playbooks permissions to certain roles:
    getPlaybooksPermissionsAddManageRoles, defined in https://github.com/mattermost/mattermost-server/blob/56a093ceaee6389a01a35b6d4626ef5a9fea4759/app/permissions_migrations.go#L1056-L1072
    The specific roles ('%playbook%') are removed in the procedure below, but the migrations also add new rows to the Systems table marking the migrations as complete.
    This in-app migration does not happen in the script, so we remove that rows here. */
DELETE FROM Systems WHERE Name = 'playbooks_manage_roles';

/* The server migration contains an in-app migration that adds boards permissions to certain roles:
   getProductsBoardsPermissions, defined in https://github.com/mattermost/mattermost-server/blob/282bd351e3767dcfd8c8340da2e0915197c0dbcb/app/permissions_migrations.go#L1074-L1093
   The specific roles (sysconsole_read_product_boards and sysconsole_write_product_boards) are removed in the procedure below,
   but the migrations also adds a new row to the Systems table marking the migrations as complete.
   This in-app migration does not happen in the script, so we remove that row here. */
DELETE FROM Systems WHERE Name = 'products_boards';

/* The server migration contains an in-app migration that adds Ids to the Teams whose InviteId is an empty string:
   doRemainingSchemaMigrations, defined in https://github.com/mattermost/mattermost-server/blob/282bd351e3767dcfd8c8340da2e0915197c0dbcb/app/migrations.go#L515-L540
   The migration is not replicated in the script, since it happens in-app, but the server adds a new row to the
   Systems table marking the table as complete, which the script doesn't do, so we remove that row here. */
DELETE FROM Systems WHERE Name = 'RemainingSchemaMigrations';

/* The server migration contains three in-app migration that adds a new role and new permissions
   related to custom groups. The migrations are:
     - doCustomGroupAdminRoleCreationMigration https://github.com/mattermost/mattermost-server/blob/282bd351e3767dcfd8c8340da2e0915197c0dbcb/app/migrations.go#L345-L469
     - getAddCustomUserGroupsPermissions https://github.com/mattermost/mattermost-server/blob/282bd351e3767dcfd8c8340da2e0915197c0dbcb/app/permissions_migrations.go#L974-L995
     - getAddCustomUserGroupsPermissionRestore https://github.com/mattermost/mattermost-server/blob/282bd351e3767dcfd8c8340da2e0915197c0dbcb/app/permissions_migrations.go#L997-L1019
   The specific roles and permissions are removed in the procedure below, but the migrations also
   adds a new row to the Roles table for the new role and new rows to the Systems table marking the
   migrations as complete.
   This in-app migration does not happen in the script, so we remove that row here. */
DELETE FROM Roles WHERE Name = 'system_custom_group_admin';
DELETE FROM Systems WHERE Name = 'CustomGroupAdminRoleCreationMigrationComplete';
DELETE FROM Systems WHERE Name = 'custom_groups_permissions';
DELETE FROM Systems WHERE Name = 'custom_groups_permission_restore';

/* The server migration contains an in-app migration that updates the config, setting ServiceSettings.PostPriority
   to true, doPostPriorityConfigDefaultTrueMigration, defined in https://github.com/mattermost/mattermost-server/blob/282bd351e3767dcfd8c8340da2e0915197c0dbcb/app/migrations.go#L542-L560
   The migration is not replicated in the script, since it happens in-app, but the server adds a new row to the
   Systems table marking the table as complete, which the script doesn't do, so we remove that row here. */
DELETE FROM Systems WHERE Name = 'PostPriorityConfigDefaultTrueMigrationComplete';

/* The rest of this script defines and executes a procedure to update the Roles table. It performs several changes:
     1. Set the UpdateAt column of all rows to a fixed value, so that the server migration changes to this column
        do not appear in the diff.
     2. Remove the set of specific permissions added in the server migration that is not covered by the script, as
        this logic happens all in-app after the normal DB migrations.
     3. Set a consistent order in the Permissions column, which is modelled a space-separated string containing each of
        the different permissions each role has. This change is the reason why we need a complex procedure, which creates
        a temporary table that pairs each single permission to its corresponding ID. So if the Roles table contains two
        rows like:
          Id: 'abcd'
          Permissions: 'view_team read_public_channel invite_user'
          Id: 'efgh'
          Permissions: 'view_team create_emojis'
        then the new temporary table will contain five rows like:
          Id: 'abcd'
          Permissions: 'view_team'
          Id: 'abcd'
          Permissions: 'read_public_channel'
          Id: 'abcd'
          Permissions: 'invite_user'
          Id: 'efgh'
          Permissions: 'view_team'
          Id: 'efgh'
          Permissions: 'create_emojis'
*/

DROP PROCEDURE IF EXISTS splitPermissions;
DROP PROCEDURE IF EXISTS sortAndFilterPermissionsInRoles;

DROP TEMPORARY TABLE IF EXISTS temp_roles;
CREATE TEMPORARY TABLE temp_roles(id varchar(26), permission longtext);

DELIMITER //

/* Auxiliary procedure that splits the space-separated permissions string into single rows that are inserted
   in the temporary temp_roles table along with their corresponding ID. */
CREATE PROCEDURE splitPermissions(
  IN id varchar(26),
  IN permissionsString longtext
)
BEGIN
  DECLARE idx INT DEFAULT 0;
  SELECT TRIM(permissionsString) INTO permissionsString;
  SELECT LOCATE(' ', permissionsString) INTO idx;
  WHILE idx > 0 DO
    INSERT INTO temp_roles SELECT id, TRIM(LEFT(permissionsString, idx));
    SELECT SUBSTR(permissionsString, idx+1) INTO permissionsString;
    SELECT LOCATE(' ', permissionsString) INTO idx;
  END WHILE;
  INSERT INTO temp_roles(id, permission) VALUES(id, TRIM(permissionsString));
END; //

/* Main procedure that does update the Roles table */
CREATE PROCEDURE sortAndFilterPermissionsInRoles()
BEGIN
  DECLARE done INT DEFAULT FALSE;
  DECLARE rolesId varchar(26) DEFAULT '';
  DECLARE rolesPermissions longtext DEFAULT '';
  DECLARE cur1 CURSOR FOR SELECT Id, Permissions FROM Roles;
  DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = TRUE;

  /* 1. Set a fixed value in the UpdateAt column for all rows in Roles table */
  UPDATE Roles SET UpdateAt = 1;

  /* Call splitPermissions for every row in the Roles table, thus populating the
     temp_roles table. */
  OPEN cur1;
  read_loop: LOOP
    FETCH cur1 INTO rolesId, rolesPermissions;
    IF done THEN
      LEAVE read_loop;
    END IF;
    CALL splitPermissions(rolesId, rolesPermissions);
  END LOOP;
  CLOSE cur1;

  /* 2. Filter out the new permissions added by the in-app migrations */
  DELETE FROM temp_roles WHERE permission LIKE 'sysconsole_read_products_boards';
  DELETE FROM temp_roles WHERE permission LIKE 'sysconsole_write_products_boards';
  DELETE FROM temp_roles WHERE permission LIKE 'playbook_public_manage_roles';
  DELETE FROM temp_roles WHERE permission LIKE 'playbook_private_manage_roles';
  DELETE FROM temp_roles WHERE permission LIKE '%custom_group%';

  /* Temporarily set to the maximum permitted value, since the call to group_concat
     below needs a value bigger than the default */
  SET group_concat_max_len = 18446744073709551615;

  /* 3. Update the Permissions column in the Roles table with the filtered, sorted permissions,
     concatenated again as a space-separated string */
  UPDATE
    Roles INNER JOIN (
      SELECT temp_roles.id as Id, TRIM(group_concat(temp_roles.permission ORDER BY temp_roles.permission SEPARATOR ' ')) as Permissions
        FROM Roles JOIN temp_roles ON Roles.Id = temp_roles.id
        GROUP BY temp_roles.id
    ) AS Sorted
    ON Roles.Id = Sorted.Id
    SET Roles.Permissions = Sorted.Permissions;

    /* Reset group_concat_max_len to its default value */
    SET group_concat_max_len = 1024;
END; //
DELIMITER ;

CALL sortAndFilterPermissionsInRoles();

DROP TEMPORARY TABLE IF EXISTS temp_roles;
