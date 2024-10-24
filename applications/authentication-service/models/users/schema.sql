-- Designing a database schema for an authorization microservice involves defining tables and relationships to manage users, roles, permissions, and the association between them. Here's a sample schema that covers these core aspects:
--
-- ### Tables and Relationships
--
-- 1. **Users**: Stores user information.
-- 2. **Roles**: Defines various roles within the system.
-- 3. **Permissions**: Defines specific permissions that can be granted.
-- 4. **UserRoles**: Maps users to roles.
-- 5. **RolePermissions**: Maps roles to permissions.
-- 6. **UserPermissions**: (Optional) Directly maps users to specific permissions, if needed.
--
-- ### Explanation
--
-- 1. **Users**: This table holds basic information about users such as username, email, and password hash. Timestamps are included to track creation and updates.
--
-- 2. **Roles**: This table holds the different roles that can be assigned to users, with a description for each role.
--
-- 3. **Permissions**: This table holds the different permissions that can be assigned to roles, with a description for each permission.
--
-- 4. **UserRoles**: This table creates a many-to-many relationship between users and roles. Each user can have multiple roles, and each role can be assigned to multiple users.
--
-- 5. **RolePermissions**: This table creates a many-to-many relationship between roles and permissions. Each role can have multiple permissions, and each permission can be assigned to multiple roles.
--
-- 6. **UserPermissions** (Optional): This table allows for direct assignment of permissions to users, bypassing roles. This can be useful for special cases where certain users need specific permissions that are not covered by their roles.
--
-- ### Considerations
--
-- - **Indexes**: Ensure that appropriate indexes are created on foreign keys and any frequently queried columns to optimize performance.
-- - **Security**: Hash and salt passwords using a strong algorithm (e.g., bcrypt) to enhance security.
-- - **Auditing**: Consider adding additional fields or tables for auditing purposes to track changes made to roles and permissions.
-- - **Scalability**: As the user base grows, consider implementing strategies for database sharding, replication, and load balancing to maintain performance.
--
-- This schema provides a solid foundation for an authorization microservice, allowing for flexible and scalable management of users, roles, and permissions.

CREATE TABLE "User"
(
    "id"           bigserial
        CONSTRAINT "user-id-primary-key" primary key,

    "email"               varchar(255)                                                                         not null
        CONSTRAINT "user-email-validation-constraint" CHECK ("User"."email" ~* '^[A-Za-z0-9._+%-]+@[A-Za-z0-9.-]+[.][A-Za-z]+$')
        CONSTRAINT "user-email-unique-constraint" unique,

    "password"     varchar(255) NOT NULL,

    "creation"     timestamp with time zone default now(),
    "modification" timestamp with time zone,
    "deletion"     timestamp with time zone
);

COMMENT ON COLUMN "User".id IS 'ID represents a PostgreSQL-generated unique identifier.';

CREATE INDEX IF NOT EXISTS "user-email-index" on "User" (email);
CREATE INDEX IF NOT EXISTS "user-deletion-index" on "User" (deletion);
