# DynamoDB Table Design for Multi-Tenant Application (Agencies and Members)

## Table Schema

- **Primary Key**
  - **Partition Key**: `PK`
  - **Sort Key**: `SK`

- **Attributes**
  - `PK` (string): Composite key containing the main identifier (e.g., agency or member)
  - `SK` (string): Composite key for entity details
  - `type`: Indicates the entity type (`Agency` or `Member`)
  - `agencyId`: Identifier for the agency
  - `created`: Timestamp of creation
  - `createdBy`: Creator ID
  - `modified`: Timestamp of last modification
  - `modifiedBy`: Modifier ID
  - `name`: Agency name (for agency items)
  - `status`: Agency status (for agency items)
  - `idpid`: Member ID (for member items)

## Item Structure

Each item in the table represents either an `Agency` or a `Member`.

### 1. Agency Items
- **PK**: `AGENCY#<agencyId>`
- **SK**: `AGENCY#<agencyId>`
- **Attributes**: `type="Agency"`, `name`, `status`, `created`, `createdBy`, `modified`, `modifiedBy`

### 2. Member Items
- **PK**: `AGENCY#<agencyId>`
- **SK**: `MEMBER#<idpid>`
- **Attributes**: `type="Member"`, `agencyId`, `idpid`, `created`, `createdBy`, `modified`, `modifiedBy`

## Global Secondary Indexes (GSIs)

To support the required access patterns, the following GSIs are added:

### **GSI1** - For Listing All Agencies
- **Partition Key**: `type`
- **Sort Key**: `created`
- **Purpose**: Allows querying all agency items by `type="Agency"`, ordered by `created` timestamp.

### **GSI2** - For Listing Members by `idpid`
- **Partition Key**: `idpid`
- **Sort Key**: `agencyId`
- **Purpose**: Enables querying members by `idpid` directly.

## Access Patterns and Query Examples

### 1. Create, Update, Delete Agencies
   - **Create/Update**: Use `PutItem` or `UpdateItem` with `PK="AGENCY#<agencyId>"` and `SK="AGENCY#<agencyId>"`
   - **Delete**: Use `DeleteItem` with `PK="AGENCY#<agencyId>"` and `SK="AGENCY#<agencyId>"`

### 2. Create, Update, Delete Members
   - **Create/Update**: Use `PutItem` or `UpdateItem` with `PK="AGENCY#<agencyId>"` and `SK="MEMBER#<idpid>"`
   - **Delete**: Use `DeleteItem` with `PK="AGENCY#<agencyId>"` and `SK="MEMBER#<idpid>"`

### 3. Read an Agency by ID
   - **Query**: `PK="AGENCY#<agencyId>"` and `SK="AGENCY#<agencyId>"`

### 4. List Members by Agency ID
   - **Query**: `PK="AGENCY#<agencyId>"` with `begins_with(SK, "MEMBER#")`

### 5. List All Agencies
   - **Query**: Using GSI1, query where `type="Agency"`

### 6. List Members by `idpid`
   - **Query**: Using GSI2, query by `idpid=<idpid>`

## Summary of Schema Design

| Field        | Notes                                                                 |
| ------------ | --------------------------------------------------------------------- |
| `PK`         | `AGENCY#<agencyId>` for both agencies and members                    |
| `SK`         | `AGENCY#<agencyId>` for agencies, `MEMBER#<idpid>` for members       |
| `GSI1`       | `type` (PK), `created` (SK) for listing all agencies                 |
| `GSI2`       | `idpid` (PK), `agencyId` (SK) for listing members by `idpid`         |
