// Code generated by lockedgen using 'go generate'. DO NOT EDIT.

// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package satellitedb

import (
	"context"
	"crypto"
	"sync"
	"time"

	"github.com/skyrings/skyring-common/tools/uuid"

	"storj.io/storj/internal/memory"
	"storj.io/storj/pkg/audit"
	"storj.io/storj/pkg/certdb"
	"storj.io/storj/pkg/datarepair/irreparable"
	"storj.io/storj/pkg/datarepair/queue"
	"storj.io/storj/pkg/macaroon"
	"storj.io/storj/pkg/overlay"
	"storj.io/storj/pkg/pb"
	"storj.io/storj/pkg/storj"
	"storj.io/storj/satellite"
	"storj.io/storj/satellite/accounting"
	"storj.io/storj/satellite/attribution"
	"storj.io/storj/satellite/console"
	"storj.io/storj/satellite/metainfo"
	"storj.io/storj/satellite/orders"
	"storj.io/storj/satellite/rewards"
)

// locked implements a locking wrapper around satellite.DB.
type locked struct {
	sync.Locker
	db satellite.DB
}

// newLocked returns database wrapped with locker.
func newLocked(db satellite.DB) satellite.DB {
	return &locked{&sync.Mutex{}, db}
}

// Attribution returns database for partner keys information
func (m *locked) Attribution() attribution.DB {
	m.Lock()
	defer m.Unlock()
	return &lockedAttribution{m.Locker, m.db.Attribution()}
}

// lockedAttribution implements locking wrapper for attribution.DB
type lockedAttribution struct {
	sync.Locker
	db attribution.DB
}

// Get retrieves attribution info using project id and bucket name.
func (m *lockedAttribution) Get(ctx context.Context, projectID uuid.UUID, bucketName []byte) (*attribution.Info, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.Get(ctx, projectID, bucketName)
}

// Insert creates and stores new Info
func (m *lockedAttribution) Insert(ctx context.Context, info *attribution.Info) (*attribution.Info, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.Insert(ctx, info)
}

// QueryAttribution queries partner bucket attribution data
func (m *lockedAttribution) QueryAttribution(ctx context.Context, partnerID uuid.UUID, start time.Time, end time.Time) ([]*attribution.CSVRow, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.QueryAttribution(ctx, partnerID, start, end)
}

// Buckets returns the database to interact with buckets
func (m *locked) Buckets() metainfo.BucketsDB {
	m.Lock()
	defer m.Unlock()
	return &lockedBuckets{m.Locker, m.db.Buckets()}
}

// lockedBuckets implements locking wrapper for metainfo.BucketsDB
type lockedBuckets struct {
	sync.Locker
	db metainfo.BucketsDB
}

// Create creates a new bucket
func (m *lockedBuckets) CreateBucket(ctx context.Context, bucket storj.Bucket) (_ storj.Bucket, err error) {
	m.Lock()
	defer m.Unlock()
	return m.db.CreateBucket(ctx, bucket)
}

// Delete deletes a bucket
func (m *lockedBuckets) DeleteBucket(ctx context.Context, bucketName []byte, projectID uuid.UUID) (err error) {
	m.Lock()
	defer m.Unlock()
	return m.db.DeleteBucket(ctx, bucketName, projectID)
}

// Get returns an existing bucket
func (m *lockedBuckets) GetBucket(ctx context.Context, bucketName []byte, projectID uuid.UUID) (bucket storj.Bucket, err error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetBucket(ctx, bucketName, projectID)
}

// List returns all buckets for a project
func (m *lockedBuckets) ListBuckets(ctx context.Context, projectID uuid.UUID, listOpts storj.BucketListOptions, allowedBuckets macaroon.AllowedBuckets) (bucketList storj.BucketList, err error) {
	m.Lock()
	defer m.Unlock()
	return m.db.ListBuckets(ctx, projectID, listOpts, allowedBuckets)
}

// CertDB returns database for storing uplink's public key & ID
func (m *locked) CertDB() certdb.DB {
	m.Lock()
	defer m.Unlock()
	return &lockedCertDB{m.Locker, m.db.CertDB()}
}

// lockedCertDB implements locking wrapper for certdb.DB
type lockedCertDB struct {
	sync.Locker
	db certdb.DB
}

// GetPublicKey gets the public key of uplink corresponding to uplink id
func (m *lockedCertDB) GetPublicKey(ctx context.Context, a1 storj.NodeID) (crypto.PublicKey, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetPublicKey(ctx, a1)
}

// SavePublicKey adds a new bandwidth agreement.
func (m *lockedCertDB) SavePublicKey(ctx context.Context, a1 storj.NodeID, a2 crypto.PublicKey) error {
	m.Lock()
	defer m.Unlock()
	return m.db.SavePublicKey(ctx, a1, a2)
}

// Close closes the database
func (m *locked) Close() error {
	m.Lock()
	defer m.Unlock()
	return m.db.Close()
}

// Console returns database for satellite console
func (m *locked) Console() console.DB {
	m.Lock()
	defer m.Unlock()
	return &lockedConsole{m.Locker, m.db.Console()}
}

// lockedConsole implements locking wrapper for console.DB
type lockedConsole struct {
	sync.Locker
	db console.DB
}

// APIKeys is a getter for APIKeys repository
func (m *lockedConsole) APIKeys() console.APIKeys {
	m.Lock()
	defer m.Unlock()
	return &lockedAPIKeys{m.Locker, m.db.APIKeys()}
}

// lockedAPIKeys implements locking wrapper for console.APIKeys
type lockedAPIKeys struct {
	sync.Locker
	db console.APIKeys
}

// Create creates and stores new APIKeyInfo
func (m *lockedAPIKeys) Create(ctx context.Context, head []byte, info console.APIKeyInfo) (*console.APIKeyInfo, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.Create(ctx, head, info)
}

// Delete deletes APIKeyInfo from store
func (m *lockedAPIKeys) Delete(ctx context.Context, id uuid.UUID) error {
	m.Lock()
	defer m.Unlock()
	return m.db.Delete(ctx, id)
}

// Get retrieves APIKeyInfo with given ID
func (m *lockedAPIKeys) Get(ctx context.Context, id uuid.UUID) (*console.APIKeyInfo, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.Get(ctx, id)
}

// GetByHead retrieves APIKeyInfo for given key head
func (m *lockedAPIKeys) GetByHead(ctx context.Context, head []byte) (*console.APIKeyInfo, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetByHead(ctx, head)
}

// GetByProjectID retrieves list of APIKeys for given projectID
func (m *lockedAPIKeys) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]console.APIKeyInfo, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetByProjectID(ctx, projectID)
}

// Update updates APIKeyInfo in store
func (m *lockedAPIKeys) Update(ctx context.Context, key console.APIKeyInfo) error {
	m.Lock()
	defer m.Unlock()
	return m.db.Update(ctx, key)
}

// BucketUsage is a getter for accounting.BucketUsage repository
func (m *lockedConsole) BucketUsage() accounting.BucketUsage {
	m.Lock()
	defer m.Unlock()
	return &lockedBucketUsage{m.Locker, m.db.BucketUsage()}
}

// lockedBucketUsage implements locking wrapper for accounting.BucketUsage
type lockedBucketUsage struct {
	sync.Locker
	db accounting.BucketUsage
}

func (m *lockedBucketUsage) Create(ctx context.Context, rollup accounting.BucketRollup) (*accounting.BucketRollup, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.Create(ctx, rollup)
}

func (m *lockedBucketUsage) Delete(ctx context.Context, id uuid.UUID) error {
	m.Lock()
	defer m.Unlock()
	return m.db.Delete(ctx, id)
}

func (m *lockedBucketUsage) Get(ctx context.Context, id uuid.UUID) (*accounting.BucketRollup, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.Get(ctx, id)
}

func (m *lockedBucketUsage) GetPaged(ctx context.Context, cursor *accounting.BucketRollupCursor) ([]accounting.BucketRollup, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetPaged(ctx, cursor)
}

// ProjectInvoiceStamps is a getter for ProjectInvoiceStamps repository
func (m *lockedConsole) ProjectInvoiceStamps() console.ProjectInvoiceStamps {
	m.Lock()
	defer m.Unlock()
	return &lockedProjectInvoiceStamps{m.Locker, m.db.ProjectInvoiceStamps()}
}

// lockedProjectInvoiceStamps implements locking wrapper for console.ProjectInvoiceStamps
type lockedProjectInvoiceStamps struct {
	sync.Locker
	db console.ProjectInvoiceStamps
}

func (m *lockedProjectInvoiceStamps) Create(ctx context.Context, stamp console.ProjectInvoiceStamp) (*console.ProjectInvoiceStamp, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.Create(ctx, stamp)
}

func (m *lockedProjectInvoiceStamps) GetAll(ctx context.Context, projectID uuid.UUID) ([]console.ProjectInvoiceStamp, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetAll(ctx, projectID)
}

func (m *lockedProjectInvoiceStamps) GetByProjectIDStartDate(ctx context.Context, projectID uuid.UUID, startDate time.Time) (*console.ProjectInvoiceStamp, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetByProjectIDStartDate(ctx, projectID, startDate)
}

// ProjectMembers is a getter for ProjectMembers repository
func (m *lockedConsole) ProjectMembers() console.ProjectMembers {
	m.Lock()
	defer m.Unlock()
	return &lockedProjectMembers{m.Locker, m.db.ProjectMembers()}
}

// lockedProjectMembers implements locking wrapper for console.ProjectMembers
type lockedProjectMembers struct {
	sync.Locker
	db console.ProjectMembers
}

// Delete is a method for deleting project member by memberID and projectID from the database.
func (m *lockedProjectMembers) Delete(ctx context.Context, memberID uuid.UUID, projectID uuid.UUID) error {
	m.Lock()
	defer m.Unlock()
	return m.db.Delete(ctx, memberID, projectID)
}

// GetByMemberID is a method for querying project members from the database by memberID.
func (m *lockedProjectMembers) GetByMemberID(ctx context.Context, memberID uuid.UUID) ([]console.ProjectMember, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetByMemberID(ctx, memberID)
}

// GetByProjectID is a method for querying project members from the database by projectID, offset and limit.
func (m *lockedProjectMembers) GetByProjectID(ctx context.Context, projectID uuid.UUID, pagination console.Pagination) ([]console.ProjectMember, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetByProjectID(ctx, projectID, pagination)
}

// Insert is a method for inserting project member into the database.
func (m *lockedProjectMembers) Insert(ctx context.Context, memberID uuid.UUID, projectID uuid.UUID) (*console.ProjectMember, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.Insert(ctx, memberID, projectID)
}

// ProjectPayments is a getter for ProjectPayments repository
func (m *lockedConsole) ProjectPayments() console.ProjectPayments {
	m.Lock()
	defer m.Unlock()
	return &lockedProjectPayments{m.Locker, m.db.ProjectPayments()}
}

// lockedProjectPayments implements locking wrapper for console.ProjectPayments
type lockedProjectPayments struct {
	sync.Locker
	db console.ProjectPayments
}

func (m *lockedProjectPayments) Create(ctx context.Context, info console.ProjectPayment) (*console.ProjectPayment, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.Create(ctx, info)
}

func (m *lockedProjectPayments) Delete(ctx context.Context, projectPaymentID uuid.UUID) error {
	m.Lock()
	defer m.Unlock()
	return m.db.Delete(ctx, projectPaymentID)
}

func (m *lockedProjectPayments) GetByID(ctx context.Context, projectPaymentID uuid.UUID) (*console.ProjectPayment, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetByID(ctx, projectPaymentID)
}

func (m *lockedProjectPayments) GetByPayerID(ctx context.Context, payerID uuid.UUID) ([]*console.ProjectPayment, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetByPayerID(ctx, payerID)
}

func (m *lockedProjectPayments) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*console.ProjectPayment, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetByProjectID(ctx, projectID)
}

func (m *lockedProjectPayments) GetDefaultByProjectID(ctx context.Context, projectID uuid.UUID) (*console.ProjectPayment, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetDefaultByProjectID(ctx, projectID)
}

func (m *lockedProjectPayments) Update(ctx context.Context, info console.ProjectPayment) error {
	m.Lock()
	defer m.Unlock()
	return m.db.Update(ctx, info)
}

// Projects is a getter for Projects repository
func (m *lockedConsole) Projects() console.Projects {
	m.Lock()
	defer m.Unlock()
	return &lockedProjects{m.Locker, m.db.Projects()}
}

// lockedProjects implements locking wrapper for console.Projects
type lockedProjects struct {
	sync.Locker
	db console.Projects
}

// Delete is a method for deleting project by Id from the database.
func (m *lockedProjects) Delete(ctx context.Context, id uuid.UUID) error {
	m.Lock()
	defer m.Unlock()
	return m.db.Delete(ctx, id)
}

// Get is a method for querying project from the database by id.
func (m *lockedProjects) Get(ctx context.Context, id uuid.UUID) (*console.Project, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.Get(ctx, id)
}

// GetAll is a method for querying all projects from the database.
func (m *lockedProjects) GetAll(ctx context.Context) ([]console.Project, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetAll(ctx)
}

// GetByUserID is a method for querying all projects from the database by userID.
func (m *lockedProjects) GetByUserID(ctx context.Context, userID uuid.UUID) ([]console.Project, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetByUserID(ctx, userID)
}

// GetCreatedBefore retrieves all projects created before provided date
func (m *lockedProjects) GetCreatedBefore(ctx context.Context, before time.Time) ([]console.Project, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetCreatedBefore(ctx, before)
}

// Insert is a method for inserting project into the database.
func (m *lockedProjects) Insert(ctx context.Context, project *console.Project) (*console.Project, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.Insert(ctx, project)
}

// Update is a method for updating project entity.
func (m *lockedProjects) Update(ctx context.Context, project *console.Project) error {
	m.Lock()
	defer m.Unlock()
	return m.db.Update(ctx, project)
}

// RegistrationTokens is a getter for RegistrationTokens repository
func (m *lockedConsole) RegistrationTokens() console.RegistrationTokens {
	m.Lock()
	defer m.Unlock()
	return &lockedRegistrationTokens{m.Locker, m.db.RegistrationTokens()}
}

// lockedRegistrationTokens implements locking wrapper for console.RegistrationTokens
type lockedRegistrationTokens struct {
	sync.Locker
	db console.RegistrationTokens
}

// Create creates new registration token
func (m *lockedRegistrationTokens) Create(ctx context.Context, projectLimit int) (*console.RegistrationToken, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.Create(ctx, projectLimit)
}

// GetByOwnerID retrieves RegTokenInfo by ownerID
func (m *lockedRegistrationTokens) GetByOwnerID(ctx context.Context, ownerID uuid.UUID) (*console.RegistrationToken, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetByOwnerID(ctx, ownerID)
}

// GetBySecret retrieves RegTokenInfo with given Secret
func (m *lockedRegistrationTokens) GetBySecret(ctx context.Context, secret console.RegistrationSecret) (*console.RegistrationToken, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetBySecret(ctx, secret)
}

// UpdateOwner updates registration token's owner
func (m *lockedRegistrationTokens) UpdateOwner(ctx context.Context, secret console.RegistrationSecret, ownerID uuid.UUID) error {
	m.Lock()
	defer m.Unlock()
	return m.db.UpdateOwner(ctx, secret, ownerID)
}

// ResetPasswordTokens is a getter for ResetPasswordTokens repository
func (m *lockedConsole) ResetPasswordTokens() console.ResetPasswordTokens {
	m.Lock()
	defer m.Unlock()
	return &lockedResetPasswordTokens{m.Locker, m.db.ResetPasswordTokens()}
}

// lockedResetPasswordTokens implements locking wrapper for console.ResetPasswordTokens
type lockedResetPasswordTokens struct {
	sync.Locker
	db console.ResetPasswordTokens
}

// Create creates new reset password token
func (m *lockedResetPasswordTokens) Create(ctx context.Context, ownerID uuid.UUID) (*console.ResetPasswordToken, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.Create(ctx, ownerID)
}

// Delete deletes ResetPasswordToken by ResetPasswordSecret
func (m *lockedResetPasswordTokens) Delete(ctx context.Context, secret console.ResetPasswordSecret) error {
	m.Lock()
	defer m.Unlock()
	return m.db.Delete(ctx, secret)
}

// GetByOwnerID retrieves ResetPasswordToken by ownerID
func (m *lockedResetPasswordTokens) GetByOwnerID(ctx context.Context, ownerID uuid.UUID) (*console.ResetPasswordToken, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetByOwnerID(ctx, ownerID)
}

// GetBySecret retrieves ResetPasswordToken with given secret
func (m *lockedResetPasswordTokens) GetBySecret(ctx context.Context, secret console.ResetPasswordSecret) (*console.ResetPasswordToken, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetBySecret(ctx, secret)
}

// UsageRollups is a getter for UsageRollups repository
func (m *lockedConsole) UsageRollups() console.UsageRollups {
	m.Lock()
	defer m.Unlock()
	return &lockedUsageRollups{m.Locker, m.db.UsageRollups()}
}

// lockedUsageRollups implements locking wrapper for console.UsageRollups
type lockedUsageRollups struct {
	sync.Locker
	db console.UsageRollups
}

func (m *lockedUsageRollups) GetBucketTotals(ctx context.Context, projectID uuid.UUID, cursor console.BucketUsageCursor, since time.Time, before time.Time) (*console.BucketUsagePage, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetBucketTotals(ctx, projectID, cursor, since, before)
}

func (m *lockedUsageRollups) GetBucketUsageRollups(ctx context.Context, projectID uuid.UUID, since time.Time, before time.Time) ([]console.BucketUsageRollup, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetBucketUsageRollups(ctx, projectID, since, before)
}

func (m *lockedUsageRollups) GetProjectTotal(ctx context.Context, projectID uuid.UUID, since time.Time, before time.Time) (*console.ProjectUsage, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetProjectTotal(ctx, projectID, since, before)
}

// UserCredits is a getter for UserCredits repository
func (m *lockedConsole) UserCredits() console.UserCredits {
	m.Lock()
	defer m.Unlock()
	return &lockedUserCredits{m.Locker, m.db.UserCredits()}
}

// lockedUserCredits implements locking wrapper for console.UserCredits
type lockedUserCredits struct {
	sync.Locker
	db console.UserCredits
}

func (m *lockedUserCredits) Create(ctx context.Context, userCredit console.UserCredit) (*console.UserCredit, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.Create(ctx, userCredit)
}

func (m *lockedUserCredits) GetCreditUsage(ctx context.Context, userID uuid.UUID, expirationEndDate time.Time) (*console.UserCreditUsage, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetCreditUsage(ctx, userID, expirationEndDate)
}

func (m *lockedUserCredits) UpdateAvailableCredits(ctx context.Context, creditsToCharge int, id uuid.UUID, billingStartDate time.Time) (remainingCharge int, err error) {
	m.Lock()
	defer m.Unlock()
	return m.db.UpdateAvailableCredits(ctx, creditsToCharge, id, billingStartDate)
}

// UserPayments is a getter for UserPayments repository
func (m *lockedConsole) UserPayments() console.UserPayments {
	m.Lock()
	defer m.Unlock()
	return &lockedUserPayments{m.Locker, m.db.UserPayments()}
}

// lockedUserPayments implements locking wrapper for console.UserPayments
type lockedUserPayments struct {
	sync.Locker
	db console.UserPayments
}

func (m *lockedUserPayments) Create(ctx context.Context, info console.UserPayment) (*console.UserPayment, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.Create(ctx, info)
}

func (m *lockedUserPayments) Get(ctx context.Context, userID uuid.UUID) (*console.UserPayment, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.Get(ctx, userID)
}

// Users is a getter for Users repository
func (m *lockedConsole) Users() console.Users {
	m.Lock()
	defer m.Unlock()
	return &lockedUsers{m.Locker, m.db.Users()}
}

// lockedUsers implements locking wrapper for console.Users
type lockedUsers struct {
	sync.Locker
	db console.Users
}

// Delete is a method for deleting user by Id from the database.
func (m *lockedUsers) Delete(ctx context.Context, id uuid.UUID) error {
	m.Lock()
	defer m.Unlock()
	return m.db.Delete(ctx, id)
}

// Get is a method for querying user from the database by id.
func (m *lockedUsers) Get(ctx context.Context, id uuid.UUID) (*console.User, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.Get(ctx, id)
}

// GetByEmail is a method for querying user by email from the database.
func (m *lockedUsers) GetByEmail(ctx context.Context, email string) (*console.User, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetByEmail(ctx, email)
}

// Insert is a method for inserting user into the database.
func (m *lockedUsers) Insert(ctx context.Context, user *console.User) (*console.User, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.Insert(ctx, user)
}

// Update is a method for updating user entity.
func (m *lockedUsers) Update(ctx context.Context, user *console.User) error {
	m.Lock()
	defer m.Unlock()
	return m.db.Update(ctx, user)
}

// Containment returns database for containment
func (m *locked) Containment() audit.Containment {
	m.Lock()
	defer m.Unlock()
	return &lockedContainment{m.Locker, m.db.Containment()}
}

// lockedContainment implements locking wrapper for audit.Containment
type lockedContainment struct {
	sync.Locker
	db audit.Containment
}

func (m *lockedContainment) Delete(ctx context.Context, nodeID storj.NodeID) (bool, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.Delete(ctx, nodeID)
}

func (m *lockedContainment) Get(ctx context.Context, nodeID storj.NodeID) (*audit.PendingAudit, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.Get(ctx, nodeID)
}

func (m *lockedContainment) IncrementPending(ctx context.Context, pendingAudit *audit.PendingAudit) error {
	m.Lock()
	defer m.Unlock()
	return m.db.IncrementPending(ctx, pendingAudit)
}

// CreateSchema sets the schema
func (m *locked) CreateSchema(schema string) error {
	m.Lock()
	defer m.Unlock()
	return m.db.CreateSchema(schema)
}

// CreateTables initializes the database
func (m *locked) CreateTables() error {
	m.Lock()
	defer m.Unlock()
	return m.db.CreateTables()
}

// DropSchema drops the schema
func (m *locked) DropSchema(schema string) error {
	m.Lock()
	defer m.Unlock()
	return m.db.DropSchema(schema)
}

// Irreparable returns database for failed repairs
func (m *locked) Irreparable() irreparable.DB {
	m.Lock()
	defer m.Unlock()
	return &lockedIrreparable{m.Locker, m.db.Irreparable()}
}

// lockedIrreparable implements locking wrapper for irreparable.DB
type lockedIrreparable struct {
	sync.Locker
	db irreparable.DB
}

// Delete removes irreparable segment info based on segmentPath.
func (m *lockedIrreparable) Delete(ctx context.Context, segmentPath []byte) error {
	m.Lock()
	defer m.Unlock()
	return m.db.Delete(ctx, segmentPath)
}

// Get returns irreparable segment info based on segmentPath.
func (m *lockedIrreparable) Get(ctx context.Context, segmentPath []byte) (*pb.IrreparableSegment, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.Get(ctx, segmentPath)
}

// GetLimited number of segments from offset
func (m *lockedIrreparable) GetLimited(ctx context.Context, limit int, offset int64) ([]*pb.IrreparableSegment, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetLimited(ctx, limit, offset)
}

// IncrementRepairAttempts increments the repair attempts.
func (m *lockedIrreparable) IncrementRepairAttempts(ctx context.Context, segmentInfo *pb.IrreparableSegment) error {
	m.Lock()
	defer m.Unlock()
	return m.db.IncrementRepairAttempts(ctx, segmentInfo)
}

// Orders returns database for orders
func (m *locked) Orders() orders.DB {
	m.Lock()
	defer m.Unlock()
	return &lockedOrders{m.Locker, m.db.Orders()}
}

// lockedOrders implements locking wrapper for orders.DB
type lockedOrders struct {
	sync.Locker
	db orders.DB
}

// CreateSerialInfo creates serial number entry in database
func (m *lockedOrders) CreateSerialInfo(ctx context.Context, serialNumber storj.SerialNumber, bucketID []byte, limitExpiration time.Time) error {
	m.Lock()
	defer m.Unlock()
	return m.db.CreateSerialInfo(ctx, serialNumber, bucketID, limitExpiration)
}

// GetBucketBandwidth gets total bucket bandwidth from period of time
func (m *lockedOrders) GetBucketBandwidth(ctx context.Context, projectID uuid.UUID, bucketName []byte, from time.Time, to time.Time) (int64, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetBucketBandwidth(ctx, projectID, bucketName, from, to)
}

// GetStorageNodeBandwidth gets total storage node bandwidth from period of time
func (m *lockedOrders) GetStorageNodeBandwidth(ctx context.Context, nodeID storj.NodeID, from time.Time, to time.Time) (int64, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetStorageNodeBandwidth(ctx, nodeID, from, to)
}

// UnuseSerialNumber removes pair serial number -> storage node id from database
func (m *lockedOrders) UnuseSerialNumber(ctx context.Context, serialNumber storj.SerialNumber, storageNodeID storj.NodeID) error {
	m.Lock()
	defer m.Unlock()
	return m.db.UnuseSerialNumber(ctx, serialNumber, storageNodeID)
}

// UpdateBucketBandwidthAllocation updates 'allocated' bandwidth for given bucket
func (m *lockedOrders) UpdateBucketBandwidthAllocation(ctx context.Context, projectID uuid.UUID, bucketName []byte, action pb.PieceAction, amount int64, intervalStart time.Time) error {
	m.Lock()
	defer m.Unlock()
	return m.db.UpdateBucketBandwidthAllocation(ctx, projectID, bucketName, action, amount, intervalStart)
}

// UpdateBucketBandwidthInline updates 'inline' bandwidth for given bucket
func (m *lockedOrders) UpdateBucketBandwidthInline(ctx context.Context, projectID uuid.UUID, bucketName []byte, action pb.PieceAction, amount int64, intervalStart time.Time) error {
	m.Lock()
	defer m.Unlock()
	return m.db.UpdateBucketBandwidthInline(ctx, projectID, bucketName, action, amount, intervalStart)
}

// UpdateBucketBandwidthSettle updates 'settled' bandwidth for given bucket
func (m *lockedOrders) UpdateBucketBandwidthSettle(ctx context.Context, projectID uuid.UUID, bucketName []byte, action pb.PieceAction, amount int64, intervalStart time.Time) error {
	m.Lock()
	defer m.Unlock()
	return m.db.UpdateBucketBandwidthSettle(ctx, projectID, bucketName, action, amount, intervalStart)
}

// UpdateStoragenodeBandwidthAllocation updates 'allocated' bandwidth for given storage nodes
func (m *lockedOrders) UpdateStoragenodeBandwidthAllocation(ctx context.Context, storageNodes []storj.NodeID, action pb.PieceAction, amount int64, intervalStart time.Time) error {
	m.Lock()
	defer m.Unlock()
	return m.db.UpdateStoragenodeBandwidthAllocation(ctx, storageNodes, action, amount, intervalStart)
}

// UpdateStoragenodeBandwidthSettle updates 'settled' bandwidth for given storage node
func (m *lockedOrders) UpdateStoragenodeBandwidthSettle(ctx context.Context, storageNode storj.NodeID, action pb.PieceAction, amount int64, intervalStart time.Time) error {
	m.Lock()
	defer m.Unlock()
	return m.db.UpdateStoragenodeBandwidthSettle(ctx, storageNode, action, amount, intervalStart)
}

// UseSerialNumber creates serial number entry in database
func (m *lockedOrders) UseSerialNumber(ctx context.Context, serialNumber storj.SerialNumber, storageNodeID storj.NodeID) ([]byte, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.UseSerialNumber(ctx, serialNumber, storageNodeID)
}

// OverlayCache returns database for caching overlay information
func (m *locked) OverlayCache() overlay.DB {
	m.Lock()
	defer m.Unlock()
	return &lockedOverlayCache{m.Locker, m.db.OverlayCache()}
}

// lockedOverlayCache implements locking wrapper for overlay.DB
type lockedOverlayCache struct {
	sync.Locker
	db overlay.DB
}

// Get looks up the node by nodeID
func (m *lockedOverlayCache) Get(ctx context.Context, nodeID storj.NodeID) (*overlay.NodeDossier, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.Get(ctx, nodeID)
}

// IsVetted returns whether or not the node reaches reputable thresholds
func (m *lockedOverlayCache) IsVetted(ctx context.Context, id storj.NodeID, criteria *overlay.NodeCriteria) (bool, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.IsVetted(ctx, id, criteria)
}

// KnownOffline filters a set of nodes to offline nodes
func (m *lockedOverlayCache) KnownOffline(ctx context.Context, a1 *overlay.NodeCriteria, a2 storj.NodeIDList) (storj.NodeIDList, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.KnownOffline(ctx, a1, a2)
}

// KnownUnreliableOrOffline filters a set of nodes to unhealth or offlines node, independent of new
func (m *lockedOverlayCache) KnownUnreliableOrOffline(ctx context.Context, a1 *overlay.NodeCriteria, a2 storj.NodeIDList) (storj.NodeIDList, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.KnownUnreliableOrOffline(ctx, a1, a2)
}

// Paginate will page through the database nodes
func (m *lockedOverlayCache) Paginate(ctx context.Context, offset int64, limit int) ([]*overlay.NodeDossier, bool, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.Paginate(ctx, offset, limit)
}

// Paginate will page through the database nodes
func (m *lockedOverlayCache) PaginateQualified(ctx context.Context, offset int64, limit int) ([]*pb.Node, bool, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.PaginateQualified(ctx, offset, limit)
}

// Reliable returns all nodes that are reliable
func (m *lockedOverlayCache) Reliable(ctx context.Context, a1 *overlay.NodeCriteria) (storj.NodeIDList, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.Reliable(ctx, a1)
}

// SelectNewStorageNodes looks up nodes based on new node criteria
func (m *lockedOverlayCache) SelectNewStorageNodes(ctx context.Context, count int, criteria *overlay.NodeCriteria) ([]*pb.Node, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.SelectNewStorageNodes(ctx, count, criteria)
}

// SelectStorageNodes looks up nodes based on criteria
func (m *lockedOverlayCache) SelectStorageNodes(ctx context.Context, count int, criteria *overlay.NodeCriteria) ([]*pb.Node, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.SelectStorageNodes(ctx, count, criteria)
}

// Update updates node address
func (m *lockedOverlayCache) UpdateAddress(ctx context.Context, value *pb.Node, defaults overlay.NodeSelectionConfig) error {
	m.Lock()
	defer m.Unlock()
	return m.db.UpdateAddress(ctx, value, defaults)
}

// UpdateNodeInfo updates node dossier with info requested from the node itself like node type, email, wallet, capacity, and version.
func (m *lockedOverlayCache) UpdateNodeInfo(ctx context.Context, node storj.NodeID, nodeInfo *pb.InfoResponse) (stats *overlay.NodeDossier, err error) {
	m.Lock()
	defer m.Unlock()
	return m.db.UpdateNodeInfo(ctx, node, nodeInfo)
}

// UpdateStats all parts of single storagenode's stats.
func (m *lockedOverlayCache) UpdateStats(ctx context.Context, request *overlay.UpdateRequest) (stats *overlay.NodeStats, err error) {
	m.Lock()
	defer m.Unlock()
	return m.db.UpdateStats(ctx, request)
}

// UpdateUptime updates a single storagenode's uptime stats.
func (m *lockedOverlayCache) UpdateUptime(ctx context.Context, nodeID storj.NodeID, isUp bool, lambda float64, weight float64, uptimeDQ float64) (stats *overlay.NodeStats, err error) {
	m.Lock()
	defer m.Unlock()
	return m.db.UpdateUptime(ctx, nodeID, isUp, lambda, weight, uptimeDQ)
}

// ProjectAccounting returns database for storing information about project data use
func (m *locked) ProjectAccounting() accounting.ProjectAccounting {
	m.Lock()
	defer m.Unlock()
	return &lockedProjectAccounting{m.Locker, m.db.ProjectAccounting()}
}

// lockedProjectAccounting implements locking wrapper for accounting.ProjectAccounting
type lockedProjectAccounting struct {
	sync.Locker
	db accounting.ProjectAccounting
}

// CreateStorageTally creates a record for BucketStorageTally in the accounting DB table
func (m *lockedProjectAccounting) CreateStorageTally(ctx context.Context, tally accounting.BucketStorageTally) error {
	m.Lock()
	defer m.Unlock()
	return m.db.CreateStorageTally(ctx, tally)
}

// GetAllocatedBandwidthTotal returns the sum of GET bandwidth usage allocated for a projectID in the past time frame
func (m *lockedProjectAccounting) GetAllocatedBandwidthTotal(ctx context.Context, projectID uuid.UUID, from time.Time) (int64, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetAllocatedBandwidthTotal(ctx, projectID, from)
}

// GetProjectUsageLimits returns project usage limit
func (m *lockedProjectAccounting) GetProjectUsageLimits(ctx context.Context, projectID uuid.UUID) (memory.Size, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetProjectUsageLimits(ctx, projectID)
}

// GetStorageTotals returns the current inline and remote storage usage for a projectID
func (m *lockedProjectAccounting) GetStorageTotals(ctx context.Context, projectID uuid.UUID) (int64, int64, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetStorageTotals(ctx, projectID)
}

// SaveTallies saves the latest project info
func (m *lockedProjectAccounting) SaveTallies(ctx context.Context, intervalStart time.Time, bucketTallies map[string]*accounting.BucketTally) ([]accounting.BucketTally, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.SaveTallies(ctx, intervalStart, bucketTallies)
}

// RepairQueue returns queue for segments that need repairing
func (m *locked) RepairQueue() queue.RepairQueue {
	m.Lock()
	defer m.Unlock()
	return &lockedRepairQueue{m.Locker, m.db.RepairQueue()}
}

// lockedRepairQueue implements locking wrapper for queue.RepairQueue
type lockedRepairQueue struct {
	sync.Locker
	db queue.RepairQueue
}

// Delete removes an injured segment.
func (m *lockedRepairQueue) Delete(ctx context.Context, s *pb.InjuredSegment) error {
	m.Lock()
	defer m.Unlock()
	return m.db.Delete(ctx, s)
}

// Insert adds an injured segment.
func (m *lockedRepairQueue) Insert(ctx context.Context, s *pb.InjuredSegment) error {
	m.Lock()
	defer m.Unlock()
	return m.db.Insert(ctx, s)
}

// Select gets an injured segment.
func (m *lockedRepairQueue) Select(ctx context.Context) (*pb.InjuredSegment, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.Select(ctx)
}

// SelectN lists limit amount of injured segments.
func (m *lockedRepairQueue) SelectN(ctx context.Context, limit int) ([]pb.InjuredSegment, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.SelectN(ctx, limit)
}

// returns database for marketing admin GUI
func (m *locked) Rewards() rewards.DB {
	m.Lock()
	defer m.Unlock()
	return &lockedRewards{m.Locker, m.db.Rewards()}
}

// lockedRewards implements locking wrapper for rewards.DB
type lockedRewards struct {
	sync.Locker
	db rewards.DB
}

func (m *lockedRewards) Create(ctx context.Context, offer *rewards.NewOffer) (*rewards.Offer, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.Create(ctx, offer)
}

func (m *lockedRewards) Finish(ctx context.Context, offerID int) error {
	m.Lock()
	defer m.Unlock()
	return m.db.Finish(ctx, offerID)
}

func (m *lockedRewards) GetCurrentByType(ctx context.Context, offerType rewards.OfferType) (*rewards.Offer, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetCurrentByType(ctx, offerType)
}

func (m *lockedRewards) ListAll(ctx context.Context) (rewards.Offers, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.ListAll(ctx)
}

// StoragenodeAccounting returns database for storing information about storagenode use
func (m *locked) StoragenodeAccounting() accounting.StoragenodeAccounting {
	m.Lock()
	defer m.Unlock()
	return &lockedStoragenodeAccounting{m.Locker, m.db.StoragenodeAccounting()}
}

// lockedStoragenodeAccounting implements locking wrapper for accounting.StoragenodeAccounting
type lockedStoragenodeAccounting struct {
	sync.Locker
	db accounting.StoragenodeAccounting
}

// DeleteTalliesBefore deletes all tallies prior to some time
func (m *lockedStoragenodeAccounting) DeleteTalliesBefore(ctx context.Context, latestRollup time.Time) error {
	m.Lock()
	defer m.Unlock()
	return m.db.DeleteTalliesBefore(ctx, latestRollup)
}

// GetBandwidthSince retrieves all bandwidth rollup entires since latestRollup
func (m *lockedStoragenodeAccounting) GetBandwidthSince(ctx context.Context, latestRollup time.Time) ([]*accounting.StoragenodeBandwidthRollup, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetBandwidthSince(ctx, latestRollup)
}

// GetTallies retrieves all tallies
func (m *lockedStoragenodeAccounting) GetTallies(ctx context.Context) ([]*accounting.StoragenodeStorageTally, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetTallies(ctx)
}

// GetTalliesSince retrieves all tallies since latestRollup
func (m *lockedStoragenodeAccounting) GetTalliesSince(ctx context.Context, latestRollup time.Time) ([]*accounting.StoragenodeStorageTally, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.GetTalliesSince(ctx, latestRollup)
}

// LastTimestamp records and returns the latest last tallied time.
func (m *lockedStoragenodeAccounting) LastTimestamp(ctx context.Context, timestampType string) (time.Time, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.LastTimestamp(ctx, timestampType)
}

// QueryNodeDailySpaceUsage returns slice of NodeSpaceUsage for given period
func (m *lockedStoragenodeAccounting) QueryNodeDailySpaceUsage(ctx context.Context, nodeID storj.NodeID, start time.Time, end time.Time) ([]accounting.NodeSpaceUsage, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.QueryNodeDailySpaceUsage(ctx, nodeID, start, end)
}

// QueryPaymentInfo queries Nodes and Accounting_Rollup on nodeID
func (m *lockedStoragenodeAccounting) QueryPaymentInfo(ctx context.Context, start time.Time, end time.Time) ([]*accounting.CSVRow, error) {
	m.Lock()
	defer m.Unlock()
	return m.db.QueryPaymentInfo(ctx, start, end)
}

// SaveRollup records tally and bandwidth rollup aggregations to the database
func (m *lockedStoragenodeAccounting) SaveRollup(ctx context.Context, latestTally time.Time, stats accounting.RollupStats) error {
	m.Lock()
	defer m.Unlock()
	return m.db.SaveRollup(ctx, latestTally, stats)
}

// SaveTallies records tallies of data at rest
func (m *lockedStoragenodeAccounting) SaveTallies(ctx context.Context, latestTally time.Time, nodeData map[storj.NodeID]float64) error {
	m.Lock()
	defer m.Unlock()
	return m.db.SaveTallies(ctx, latestTally, nodeData)
}
