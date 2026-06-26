package showroom_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"infiour.local/dms-api-server/internal/modules/showroom"
)

// ─── Mocks ───────────────────────────────────────────────────────────────────

// mockShowroomRepo satisfies showroom's internal showroomRepo interface via its exported methods.
type mockShowroomRepo struct {
	mock.Mock
}

func (m *mockShowroomRepo) CreateWithOwner(ctx context.Context, userID uint64, s *showroom.Showroom) (*showroom.Showroom, error) {
	args := m.Called(ctx, userID, s)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*showroom.Showroom), args.Error(1)
}

func (m *mockShowroomRepo) UpdateFilePaths(ctx context.Context, showroomID uint64, logoPath, bannerPath *string) error {
	args := m.Called(ctx, showroomID, logoPath, bannerPath)
	return args.Error(0)
}

func (m *mockShowroomRepo) AddMember(ctx context.Context, showroomID, targetUserID uint64, roleType string) error {
	args := m.Called(ctx, showroomID, targetUserID, roleType)
	return args.Error(0)
}

func (m *mockShowroomRepo) ListMembers(ctx context.Context, showroomID uint64, page, limit int) ([]showroom.MemberRecord, int64, error) {
	args := m.Called(ctx, showroomID, page, limit)
	return args.Get(0).([]showroom.MemberRecord), args.Get(1).(int64), args.Error(2)
}

func (m *mockShowroomRepo) GetMemberRole(ctx context.Context, showroomID, targetUserID uint64) (string, error) {
	args := m.Called(ctx, showroomID, targetUserID)
	return args.String(0), args.Error(1)
}

func (m *mockShowroomRepo) RemoveMember(ctx context.Context, showroomID, targetUserID uint64) error {
	args := m.Called(ctx, showroomID, targetUserID)
	return args.Error(0)
}

func (m *mockShowroomRepo) UpdateMemberRole(ctx context.Context, showroomID, targetUserID uint64, newRoleType string) error {
	args := m.Called(ctx, showroomID, targetUserID, newRoleType)
	return args.Error(0)
}

func (m *mockShowroomRepo) GetByID(ctx context.Context, showroomID uint64) (*showroom.Showroom, error) {
	args := m.Called(ctx, showroomID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*showroom.Showroom), args.Error(1)
}

func (m *mockShowroomRepo) UpdateShowroomFields(ctx context.Context, showroomID uint64, updates map[string]any) error {
	args := m.Called(ctx, showroomID, updates)
	return args.Error(0)
}

// mockStorageProvider satisfies storage.Provider.
type mockStorageProvider struct {
	mock.Mock
}

func (m *mockStorageProvider) Upload(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	args := m.Called(ctx, key, data, contentType)
	return args.String(0), args.Error(1)
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func makeFileHeader(filename string, size int64) *multipart.FileHeader {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	h.Set("Content-Type", "image/jpeg")
	return &multipart.FileHeader{
		Filename: filename,
		Header:   h,
		Size:     size,
	}
}

func inMemoryOpener(content []byte) func(*multipart.FileHeader) (io.ReadCloser, error) {
	return func(_ *multipart.FileHeader) (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(content)), nil
	}
}

func errorOpener(err error) func(*multipart.FileHeader) (io.ReadCloser, error) {
	return func(_ *multipart.FileHeader) (io.ReadCloser, error) {
		return nil, err
	}
}

func parsedFileHeader(t *testing.T, fieldName, filename string, content []byte) *multipart.FileHeader {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile(fieldName, filename)
	require.NoError(t, err)
	_, err = part.Write(content)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	req, err := http.NewRequest(http.MethodPost, "/", &body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", w.FormDataContentType())
	require.NoError(t, req.ParseMultipartForm(10<<20))
	return req.MultipartForm.File[fieldName][0]
}

// ─── CreateShowroom ───────────────────────────────────────────────────────────

func TestCreateShowroom_EmptyName(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	svc := showroom.NewService(repo, storage)

	_, err := svc.CreateShowroom(context.Background(), 1, &showroom.CreateShowroomRequest{Name: "  "}, nil, nil)
	assert.Error(t, err)
	repo.AssertNotCalled(t, "CreateWithOwner")
}

func TestCreateShowroom_InvalidGeolocationJSON(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	svc := showroom.NewService(repo, storage)

	_, err := svc.CreateShowroom(context.Background(), 1, &showroom.CreateShowroomRequest{
		Name:        "Showroom",
		Geolocation: "not-json",
	}, nil, nil)
	assert.Error(t, err)
	repo.AssertNotCalled(t, "CreateWithOwner")
}

func TestCreateShowroom_LogoTooLarge(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	svc := showroom.NewService(repo, storage)

	logo := makeFileHeader("logo.jpg", 11*1024*1024)

	_, err := svc.CreateShowroom(context.Background(), 1, &showroom.CreateShowroomRequest{Name: "X"}, logo, nil)
	assert.Error(t, err)
	repo.AssertNotCalled(t, "CreateWithOwner")
}

func TestCreateShowroom_BannerInvalidExtension(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	svc := showroom.NewService(repo, storage)

	banner := makeFileHeader("banner.gif", 100)

	_, err := svc.CreateShowroom(context.Background(), 1, &showroom.CreateShowroomRequest{Name: "X"}, nil, banner)
	assert.Error(t, err)
	repo.AssertNotCalled(t, "CreateWithOwner")
}

func TestCreateShowroom_CreateWithOwnerError(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	svc := showroom.NewService(repo, storage)

	repo.On("CreateWithOwner", mock.Anything, uint64(1), mock.Anything).
		Return(nil, errors.New("db error"))

	_, err := svc.CreateShowroom(context.Background(), 1, &showroom.CreateShowroomRequest{Name: "X"}, nil, nil)
	assert.Error(t, err)
	repo.AssertExpectations(t)
}

func TestCreateShowroom_NoFiles_Success(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	svc := showroom.NewService(repo, storage)

	created := &showroom.Showroom{ID: 7, Name: "MyShowroom"}
	repo.On("CreateWithOwner", mock.Anything, uint64(1), mock.Anything).Return(created, nil)

	resp, err := svc.CreateShowroom(context.Background(), 1, &showroom.CreateShowroomRequest{Name: "MyShowroom"}, nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint64(7), resp.ID)
	assert.Equal(t, "MyShowroom", resp.Name)
	assert.Nil(t, resp.ShowroomLogo)
	assert.Nil(t, resp.ShowroomBanner)
	repo.AssertExpectations(t)
	storage.AssertNotCalled(t, "Upload")
}

func TestCreateShowroom_WithValidGeolocation_Success(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	svc := showroom.NewService(repo, storage)

	created := &showroom.Showroom{ID: 8, Name: "Geo"}
	repo.On("CreateWithOwner", mock.Anything, uint64(2), mock.Anything).Return(created, nil)

	geo := `{"address":"123 Main St","city":"Bengaluru"}`
	resp, err := svc.CreateShowroom(context.Background(), 2, &showroom.CreateShowroomRequest{
		Name:        "Geo",
		Geolocation: geo,
	}, nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	repo.AssertExpectations(t)
}

func TestCreateShowroom_LogoOpenError_SkipsUpload(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	svc := showroom.NewService(repo, storage,
		showroom.WithFileOpener(errorOpener(errors.New("open failed"))))

	created := &showroom.Showroom{ID: 3, Name: "X"}
	repo.On("CreateWithOwner", mock.Anything, uint64(1), mock.Anything).Return(created, nil)

	logo := makeFileHeader("logo.jpg", 100)
	resp, err := svc.CreateShowroom(context.Background(), 1, &showroom.CreateShowroomRequest{Name: "X"}, logo, nil)
	assert.NoError(t, err)
	assert.Nil(t, resp.ShowroomLogo)
	storage.AssertNotCalled(t, "Upload")
}

func TestCreateShowroom_LogoStorageError_SkipsUpload(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	imgData := []byte("fake-image-data")
	svc := showroom.NewService(repo, storage,
		showroom.WithFileOpener(inMemoryOpener(imgData)))

	created := &showroom.Showroom{ID: 4, Name: "X"}
	repo.On("CreateWithOwner", mock.Anything, uint64(1), mock.Anything).Return(created, nil)

	storage.On("Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("", errors.New("upload failed"))

	logo := makeFileHeader("logo.jpg", int64(len(imgData)))
	resp, err := svc.CreateShowroom(context.Background(), 1, &showroom.CreateShowroomRequest{Name: "X"}, logo, nil)
	assert.NoError(t, err)
	assert.Nil(t, resp.ShowroomLogo)
}

func TestCreateShowroom_BothFiles_Success(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	imgData := []byte("image-bytes")
	svc := showroom.NewService(repo, storage,
		showroom.WithFileOpener(inMemoryOpener(imgData)))

	created := &showroom.Showroom{ID: 5, Name: "Both"}
	repo.On("CreateWithOwner", mock.Anything, uint64(1), mock.Anything).Return(created, nil)

	logoPath := "1/5/logo.jpg"
	bannerPath := "1/5/banner.png"
	storage.On("Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(logoPath, nil).Once()
	storage.On("Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(bannerPath, nil).Once()
	repo.On("UpdateFilePaths", mock.Anything, uint64(5), mock.Anything, mock.Anything).Return(nil)

	logo := makeFileHeader("logo.jpg", int64(len(imgData)))
	banner := makeFileHeader("banner.png", int64(len(imgData)))

	resp, err := svc.CreateShowroom(context.Background(), 1, &showroom.CreateShowroomRequest{Name: "Both"}, logo, banner)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.ShowroomLogo)
	assert.NotNil(t, resp.ShowroomBanner)
	repo.AssertExpectations(t)
	storage.AssertExpectations(t)
}

func TestCreateShowroom_BannerStorageError_SkipsUpload(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	imgData := []byte("image-bytes")
	svc := showroom.NewService(repo, storage,
		showroom.WithFileOpener(inMemoryOpener(imgData)))

	created := &showroom.Showroom{ID: 6, Name: "X"}
	repo.On("CreateWithOwner", mock.Anything, uint64(1), mock.Anything).Return(created, nil)
	storage.On("Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("", errors.New("upload failed"))

	banner := makeFileHeader("banner.jpg", int64(len(imgData)))
	resp, err := svc.CreateShowroom(context.Background(), 1, &showroom.CreateShowroomRequest{Name: "X"}, nil, banner)
	assert.NoError(t, err)
	assert.Nil(t, resp.ShowroomBanner)
}

func TestCreateShowroom_LogoOnlyUploadSuccess_UpdateFilePaths(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	imgData := []byte("img")
	svc := showroom.NewService(repo, storage,
		showroom.WithFileOpener(inMemoryOpener(imgData)))

	created := &showroom.Showroom{ID: 9, Name: "Logo"}
	repo.On("CreateWithOwner", mock.Anything, uint64(1), mock.Anything).Return(created, nil)
	storage.On("Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("1/9/logo.jpg", nil)
	repo.On("UpdateFilePaths", mock.Anything, uint64(9), mock.Anything, (*string)(nil)).Return(nil)

	logo := makeFileHeader("logo.jpg", int64(len(imgData)))
	resp, err := svc.CreateShowroom(context.Background(), 1, &showroom.CreateShowroomRequest{Name: "Logo"}, logo, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp.ShowroomLogo)
	assert.Nil(t, resp.ShowroomBanner)
	repo.AssertExpectations(t)
}

func TestCreateShowroom_DefaultOpenFile_UploadSuccess(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	svc := showroom.NewService(repo, storage)

	created := &showroom.Showroom{ID: 20, Name: "Default"}
	repo.On("CreateWithOwner", mock.Anything, uint64(1), mock.Anything).Return(created, nil)
	storage.On("Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("1/20/logo.jpg", nil)
	repo.On("UpdateFilePaths", mock.Anything, uint64(20), mock.Anything, (*string)(nil)).Return(nil)

	logo := parsedFileHeader(t, "showroom_logo", "logo.jpg", []byte("img-data"))
	resp, err := svc.CreateShowroom(context.Background(), 1, &showroom.CreateShowroomRequest{Name: "Default"}, logo, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.ShowroomLogo)
	repo.AssertExpectations(t)
	storage.AssertExpectations(t)
}

func TestCreateShowroom_EmptyContentType_UsesOctetStream(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	imgData := []byte("img")
	svc := showroom.NewService(repo, storage,
		showroom.WithFileOpener(inMemoryOpener(imgData)))

	created := &showroom.Showroom{ID: 10, Name: "CT"}
	repo.On("CreateWithOwner", mock.Anything, uint64(1), mock.Anything).Return(created, nil)
	storage.On("Upload", mock.Anything, mock.Anything, mock.Anything, "application/octet-stream").
		Return("1/10/logo.jpg", nil)
	repo.On("UpdateFilePaths", mock.Anything, uint64(10), mock.Anything, (*string)(nil)).Return(nil)

	fh := &multipart.FileHeader{
		Filename: "logo.jpg",
		Header:   make(textproto.MIMEHeader),
		Size:     int64(len(imgData)),
	}
	resp, err := svc.CreateShowroom(context.Background(), 1, &showroom.CreateShowroomRequest{Name: "CT"}, fh, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	storage.AssertExpectations(t)
}

// ─── AddMember ────────────────────────────────────────────────────────────────

func ownerRoles(showroomID uint64) map[uint64]string {
	return map[uint64]string{showroomID: "owner"}
}

func managerRoles(showroomID uint64) map[uint64]string {
	return map[uint64]string{showroomID: "manager"}
}

func employeeRoles(showroomID uint64) map[uint64]string {
	return map[uint64]string{showroomID: "employee"}
}

func TestAddMember_CallerNotMember_Forbidden(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	_, err := svc.AddMember(context.Background(), map[uint64]string{}, uint64(1), &showroom.AddMemberRequest{UserID: 99, Role: "employee"})
	assert.Error(t, err)
}

func TestAddMember_CallerIsEmployee_Forbidden(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	_, err := svc.AddMember(context.Background(), employeeRoles(1), uint64(1), &showroom.AddMemberRequest{UserID: 99, Role: "employee"})
	assert.Error(t, err)
}

func TestAddMember_InvalidRole(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	_, err := svc.AddMember(context.Background(), ownerRoles(1), uint64(1), &showroom.AddMemberRequest{UserID: 99, Role: "owner"})
	assert.Error(t, err)
	repo.AssertNotCalled(t, "AddMember")
}

func TestAddMember_ManagerTriesToAddManager_Forbidden(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	_, err := svc.AddMember(context.Background(), managerRoles(1), uint64(1), &showroom.AddMemberRequest{UserID: 99, Role: "manager"})
	assert.Error(t, err)
	repo.AssertNotCalled(t, "AddMember")
}

func TestAddMember_TargetUserNotFound(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	repo.On("AddMember", mock.Anything, uint64(1), uint64(99), "employee").
		Return(showroom.ErrTargetUserNotFound)

	_, err := svc.AddMember(context.Background(), ownerRoles(1), uint64(1), &showroom.AddMemberRequest{UserID: 99, Role: "employee"})
	assert.Error(t, err)
}

func TestAddMember_AlreadyAMember(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	repo.On("AddMember", mock.Anything, uint64(1), uint64(99), "employee").
		Return(showroom.ErrDuplicateMember)

	_, err := svc.AddMember(context.Background(), ownerRoles(1), uint64(1), &showroom.AddMemberRequest{UserID: 99, Role: "employee"})
	assert.Error(t, err)
}

func TestAddMember_RepoError(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	repo.On("AddMember", mock.Anything, uint64(1), uint64(99), "employee").
		Return(errors.New("db error"))

	_, err := svc.AddMember(context.Background(), ownerRoles(1), uint64(1), &showroom.AddMemberRequest{UserID: 99, Role: "employee"})
	assert.Error(t, err)
}

func TestAddMember_OwnerAddsEmployee_Success(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	repo.On("AddMember", mock.Anything, uint64(1), uint64(99), "employee").Return(nil)

	resp, err := svc.AddMember(context.Background(), ownerRoles(1), uint64(1), &showroom.AddMemberRequest{UserID: 99, Role: "employee"})
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), resp.ShowroomID)
	assert.Equal(t, uint64(99), resp.UserID)
	assert.Equal(t, "employee", resp.Role)
}

func TestAddMember_OwnerAddsManager_Success(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	repo.On("AddMember", mock.Anything, uint64(1), uint64(99), "manager").Return(nil)

	resp, err := svc.AddMember(context.Background(), ownerRoles(1), uint64(1), &showroom.AddMemberRequest{UserID: 99, Role: "manager"})
	assert.NoError(t, err)
	assert.Equal(t, "manager", resp.Role)
}

func TestAddMember_ManagerAddsEmployee_Success(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	repo.On("AddMember", mock.Anything, uint64(1), uint64(99), "employee").Return(nil)

	resp, err := svc.AddMember(context.Background(), managerRoles(1), uint64(1), &showroom.AddMemberRequest{UserID: 99, Role: "employee"})
	assert.NoError(t, err)
	assert.Equal(t, "employee", resp.Role)
}

// ─── ListMembers ──────────────────────────────────────────────────────────────

func TestListMembers_CallerNotMember_Forbidden(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	_, err := svc.ListMembers(context.Background(), map[uint64]string{}, uint64(1), 1, 20)
	assert.Error(t, err)
}

func TestListMembers_CallerIsEmployee_Forbidden(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	_, err := svc.ListMembers(context.Background(), employeeRoles(1), uint64(1), 1, 20)
	assert.Error(t, err)
}

func TestListMembers_RepoError(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	repo.On("ListMembers", mock.Anything, uint64(1), 1, 20).
		Return([]showroom.MemberRecord{}, int64(0), errors.New("db error"))

	_, err := svc.ListMembers(context.Background(), ownerRoles(1), uint64(1), 1, 20)
	assert.Error(t, err)
}

func TestListMembers_EmptyList(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	repo.On("ListMembers", mock.Anything, uint64(1), 1, 20).
		Return([]showroom.MemberRecord{}, int64(0), nil)

	resp, err := svc.ListMembers(context.Background(), ownerRoles(1), uint64(1), 1, 20)
	assert.NoError(t, err)
	assert.Empty(t, resp.Members)
	assert.Equal(t, int64(0), resp.Total)
}

func TestListMembers_WithNameAndPhone_Success(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	records := []showroom.MemberRecord{
		{UserID: 10, Name: "Alice", CountryCode: "+91", PhoneNumber: "9999999999", Role: "manager"},
	}
	repo.On("ListMembers", mock.Anything, uint64(1), 1, 20).
		Return(records, int64(1), nil)

	resp, err := svc.ListMembers(context.Background(), ownerRoles(1), uint64(1), 1, 20)
	assert.NoError(t, err)
	require.Len(t, resp.Members, 1)
	assert.NotNil(t, resp.Members[0].Name)
	assert.Equal(t, "Alice", *resp.Members[0].Name)
	assert.NotNil(t, resp.Members[0].PhoneNumber)
	assert.Equal(t, "+919999999999", *resp.Members[0].PhoneNumber)
}

func TestListMembers_EmptyNameAndPhone_NilFields(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	records := []showroom.MemberRecord{
		{UserID: 11, Name: "", CountryCode: "", PhoneNumber: "", Role: "employee"},
	}
	repo.On("ListMembers", mock.Anything, uint64(1), 1, 20).
		Return(records, int64(1), nil)

	resp, err := svc.ListMembers(context.Background(), managerRoles(1), uint64(1), 1, 20)
	assert.NoError(t, err)
	require.Len(t, resp.Members, 1)
	assert.Nil(t, resp.Members[0].Name)
	assert.Nil(t, resp.Members[0].PhoneNumber)
}

// ─── RemoveMember ─────────────────────────────────────────────────────────────

func TestRemoveMember_NotMemberAndNotSelf_Forbidden(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	err := svc.RemoveMember(context.Background(), uint64(1), map[uint64]string{}, uint64(1), uint64(99))
	assert.Error(t, err)
}

func TestRemoveMember_SelfRemoval_NotMemberAtAll_Forbidden(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	// callerUserID == targetUserID but no role in showroom
	err := svc.RemoveMember(context.Background(), uint64(99), map[uint64]string{}, uint64(1), uint64(99))
	assert.Error(t, err)
}

func TestRemoveMember_SelfRemoval_Success(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	repo.On("RemoveMember", mock.Anything, uint64(1), uint64(99)).Return(nil)

	err := svc.RemoveMember(context.Background(), uint64(99), managerRoles(1), uint64(1), uint64(99))
	assert.NoError(t, err)
}

func TestRemoveMember_OwnerRemovesManager_Success(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	repo.On("RemoveMember", mock.Anything, uint64(1), uint64(99)).Return(nil)

	err := svc.RemoveMember(context.Background(), uint64(1), ownerRoles(1), uint64(1), uint64(99))
	assert.NoError(t, err)
}

func TestRemoveMember_ManagerRemovesEmployee_Success(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	repo.On("GetMemberRole", mock.Anything, uint64(1), uint64(99)).Return("employee", nil)
	repo.On("RemoveMember", mock.Anything, uint64(1), uint64(99)).Return(nil)

	err := svc.RemoveMember(context.Background(), uint64(1), managerRoles(1), uint64(1), uint64(99))
	assert.NoError(t, err)
}

func TestRemoveMember_ManagerTriesToRemoveManager_Forbidden(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	repo.On("GetMemberRole", mock.Anything, uint64(1), uint64(99)).Return("manager", nil)

	err := svc.RemoveMember(context.Background(), uint64(1), managerRoles(1), uint64(1), uint64(99))
	assert.Error(t, err)
}

func TestRemoveMember_ManagerTriesToRemoveOwner_Forbidden(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	repo.On("GetMemberRole", mock.Anything, uint64(1), uint64(99)).Return("owner", nil)

	err := svc.RemoveMember(context.Background(), uint64(1), managerRoles(1), uint64(1), uint64(99))
	assert.Error(t, err)
}

func TestRemoveMember_GetMemberRole_NotFound(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	repo.On("GetMemberRole", mock.Anything, uint64(1), uint64(99)).Return("", showroom.ErrMemberNotFound)

	err := svc.RemoveMember(context.Background(), uint64(1), managerRoles(1), uint64(1), uint64(99))
	assert.Error(t, err)
}

func TestRemoveMember_GetMemberRole_DBError(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	repo.On("GetMemberRole", mock.Anything, uint64(1), uint64(99)).Return("", errors.New("db error"))

	err := svc.RemoveMember(context.Background(), uint64(1), managerRoles(1), uint64(1), uint64(99))
	assert.Error(t, err)
}

func TestRemoveMember_MemberNotFound(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	repo.On("RemoveMember", mock.Anything, uint64(1), uint64(99)).Return(showroom.ErrMemberNotFound)

	err := svc.RemoveMember(context.Background(), uint64(1), ownerRoles(1), uint64(1), uint64(99))
	assert.Error(t, err)
}

func TestRemoveMember_RepoError(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	repo.On("RemoveMember", mock.Anything, uint64(1), uint64(99)).Return(errors.New("db error"))

	err := svc.RemoveMember(context.Background(), uint64(1), ownerRoles(1), uint64(1), uint64(99))
	assert.Error(t, err)
}

// ─── UpdateMemberRole ─────────────────────────────────────────────────────────

func TestUpdateMemberRole_CallerNotOwner_Forbidden(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	_, err := svc.UpdateMemberRole(context.Background(), uint64(1), managerRoles(1), uint64(1), uint64(99), &showroom.UpdateMemberRoleRequest{Role: "employee"})
	assert.Error(t, err)
}

func TestUpdateMemberRole_SelfRoleChange_Forbidden(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	// callerUserID == targetUserID
	_, err := svc.UpdateMemberRole(context.Background(), uint64(99), ownerRoles(1), uint64(1), uint64(99), &showroom.UpdateMemberRoleRequest{Role: "employee"})
	assert.Error(t, err)
}

func TestUpdateMemberRole_InvalidRole(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	_, err := svc.UpdateMemberRole(context.Background(), uint64(1), ownerRoles(1), uint64(1), uint64(99), &showroom.UpdateMemberRoleRequest{Role: "owner"})
	assert.Error(t, err)
	repo.AssertNotCalled(t, "UpdateMemberRole")
}

func TestServiceUpdateMemberRole_MemberNotFound(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	repo.On("UpdateMemberRole", mock.Anything, uint64(1), uint64(99), "manager").
		Return(showroom.ErrMemberNotFound)

	_, err := svc.UpdateMemberRole(context.Background(), uint64(1), ownerRoles(1), uint64(1), uint64(99), &showroom.UpdateMemberRoleRequest{Role: "manager"})
	assert.Error(t, err)
}

func TestUpdateMemberRole_RepoError(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	repo.On("UpdateMemberRole", mock.Anything, uint64(1), uint64(99), "manager").
		Return(errors.New("db error"))

	_, err := svc.UpdateMemberRole(context.Background(), uint64(1), ownerRoles(1), uint64(1), uint64(99), &showroom.UpdateMemberRoleRequest{Role: "manager"})
	assert.Error(t, err)
}

func TestServiceUpdateMemberRole_Success(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	repo.On("UpdateMemberRole", mock.Anything, uint64(1), uint64(99), "manager").Return(nil)

	resp, err := svc.UpdateMemberRole(context.Background(), uint64(1), ownerRoles(1), uint64(1), uint64(99), &showroom.UpdateMemberRoleRequest{Role: "manager"})
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), resp.ShowroomID)
	assert.Equal(t, uint64(99), resp.UserID)
	assert.Equal(t, "manager", resp.Role)
}

// ─── UpdateShowroom ───────────────────────────────────────────────────────────

func existingShowroom() *showroom.Showroom {
	logo := "old/logo.jpg"
	banner := "old/banner.jpg"
	return &showroom.Showroom{ID: 1, Name: "Old Name", ShowroomLogo: &logo, ShowroomBanner: &banner}
}

func TestUpdateShowroom_Forbidden(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	_, err := svc.UpdateShowroom(context.Background(), uint64(1), employeeRoles(1), uint64(1), &showroom.UpdateShowroomRequest{}, nil, nil)
	assert.Error(t, err)
	repo.AssertNotCalled(t, "GetByID")
}

func TestUpdateShowroom_ShowroomNotFound(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	repo.On("GetByID", mock.Anything, uint64(1)).Return(nil, showroom.ErrShowroomNotFound)

	_, err := svc.UpdateShowroom(context.Background(), uint64(1), ownerRoles(1), uint64(1), &showroom.UpdateShowroomRequest{}, nil, nil)
	assert.Error(t, err)
}

func TestUpdateShowroom_GetByIDDBError(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	repo.On("GetByID", mock.Anything, uint64(1)).Return(nil, errors.New("db error"))

	_, err := svc.UpdateShowroom(context.Background(), uint64(1), ownerRoles(1), uint64(1), &showroom.UpdateShowroomRequest{}, nil, nil)
	assert.Error(t, err)
}

func TestUpdateShowroom_InvalidGeolocationJSON(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	repo.On("GetByID", mock.Anything, uint64(1)).Return(existingShowroom(), nil)

	_, err := svc.UpdateShowroom(context.Background(), uint64(1), ownerRoles(1), uint64(1),
		&showroom.UpdateShowroomRequest{Name: "New", Geolocation: "not-json"}, nil, nil)
	assert.Error(t, err)
	repo.AssertNotCalled(t, "UpdateShowroomFields")
}

func TestUpdateShowroom_LogoFileTooLarge(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	repo.On("GetByID", mock.Anything, uint64(1)).Return(existingShowroom(), nil)
	logo := makeFileHeader("logo.jpg", 11*1024*1024)

	_, err := svc.UpdateShowroom(context.Background(), uint64(1), ownerRoles(1), uint64(1),
		&showroom.UpdateShowroomRequest{}, logo, nil)
	assert.Error(t, err)
	repo.AssertNotCalled(t, "UpdateShowroomFields")
}

func TestUpdateShowroom_BannerInvalidExt(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	repo.On("GetByID", mock.Anything, uint64(1)).Return(existingShowroom(), nil)
	banner := makeFileHeader("banner.gif", 100)

	_, err := svc.UpdateShowroom(context.Background(), uint64(1), ownerRoles(1), uint64(1),
		&showroom.UpdateShowroomRequest{}, nil, banner)
	assert.Error(t, err)
	repo.AssertNotCalled(t, "UpdateShowroomFields")
}

func TestUpdateShowroom_NoChanges_Success(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	existing := existingShowroom()
	repo.On("GetByID", mock.Anything, uint64(1)).Return(existing, nil)

	resp, err := svc.UpdateShowroom(context.Background(), uint64(1), ownerRoles(1), uint64(1),
		&showroom.UpdateShowroomRequest{}, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, "Old Name", resp.Name)
	repo.AssertNotCalled(t, "UpdateShowroomFields")
}

func TestUpdateShowroom_UpdateNameAndGeo_Success(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	repo.On("GetByID", mock.Anything, uint64(1)).Return(existingShowroom(), nil)
	repo.On("UpdateShowroomFields", mock.Anything, uint64(1), mock.Anything).Return(nil)

	resp, err := svc.UpdateShowroom(context.Background(), uint64(1), ownerRoles(1), uint64(1),
		&showroom.UpdateShowroomRequest{Name: "New Name", Geolocation: `{"city":"Delhi"}`}, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, "New Name", resp.Name)
	assert.NotNil(t, resp.Geolocation)
}

func TestUpdateShowroom_RemoveLogo_Success(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	repo.On("GetByID", mock.Anything, uint64(1)).Return(existingShowroom(), nil)
	repo.On("UpdateShowroomFields", mock.Anything, uint64(1), mock.Anything).Return(nil)

	resp, err := svc.UpdateShowroom(context.Background(), uint64(1), ownerRoles(1), uint64(1),
		&showroom.UpdateShowroomRequest{RemoveLogo: "true"}, nil, nil)
	assert.NoError(t, err)
	assert.Nil(t, resp.ShowroomLogo)
}

func TestUpdateShowroom_RemoveBanner_Success(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	repo.On("GetByID", mock.Anything, uint64(1)).Return(existingShowroom(), nil)
	repo.On("UpdateShowroomFields", mock.Anything, uint64(1), mock.Anything).Return(nil)

	resp, err := svc.UpdateShowroom(context.Background(), uint64(1), ownerRoles(1), uint64(1),
		&showroom.UpdateShowroomRequest{RemoveBanner: "true"}, nil, nil)
	assert.NoError(t, err)
	assert.Nil(t, resp.ShowroomBanner)
}

func TestUpdateShowroom_LogoUploadSuccess(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	svc := showroom.NewService(repo, storage, showroom.WithFileOpener(inMemoryOpener([]byte("img"))))

	repo.On("GetByID", mock.Anything, uint64(1)).Return(existingShowroom(), nil)
	storage.On("Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("new/logo.jpg", nil)
	repo.On("UpdateShowroomFields", mock.Anything, uint64(1), mock.Anything).Return(nil)

	logo := makeFileHeader("logo.jpg", 100)
	resp, err := svc.UpdateShowroom(context.Background(), uint64(1), ownerRoles(1), uint64(1),
		&showroom.UpdateShowroomRequest{}, logo, nil)
	assert.NoError(t, err)
	require.NotNil(t, resp.ShowroomLogo)
	assert.Equal(t, "new/logo.jpg", *resp.ShowroomLogo)
}

func TestUpdateShowroom_LogoUploadFails_LogoUnchanged(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	svc := showroom.NewService(repo, storage, showroom.WithFileOpener(errorOpener(errors.New("open fail"))))

	existing := existingShowroom()
	repo.On("GetByID", mock.Anything, uint64(1)).Return(existing, nil)

	logo := makeFileHeader("logo.jpg", 100)
	resp, err := svc.UpdateShowroom(context.Background(), uint64(1), ownerRoles(1), uint64(1),
		&showroom.UpdateShowroomRequest{}, logo, nil)
	assert.NoError(t, err)
	// logo unchanged — existing logo still set
	assert.Equal(t, existing.ShowroomLogo, resp.ShowroomLogo)
	repo.AssertNotCalled(t, "UpdateShowroomFields")
}

func TestUpdateShowroom_LogoUploadOverridesRemoveLogo(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	svc := showroom.NewService(repo, storage, showroom.WithFileOpener(inMemoryOpener([]byte("img"))))

	repo.On("GetByID", mock.Anything, uint64(1)).Return(existingShowroom(), nil)
	storage.On("Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("new/logo.jpg", nil)
	repo.On("UpdateShowroomFields", mock.Anything, uint64(1), mock.Anything).Return(nil)

	logo := makeFileHeader("logo.jpg", 100)
	resp, err := svc.UpdateShowroom(context.Background(), uint64(1), ownerRoles(1), uint64(1),
		&showroom.UpdateShowroomRequest{RemoveLogo: "true"}, logo, nil)
	assert.NoError(t, err)
	require.NotNil(t, resp.ShowroomLogo)
	assert.Equal(t, "new/logo.jpg", *resp.ShowroomLogo)
}

func TestUpdateShowroom_BannerUploadSuccess(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	svc := showroom.NewService(repo, storage, showroom.WithFileOpener(inMemoryOpener([]byte("img"))))

	repo.On("GetByID", mock.Anything, uint64(1)).Return(existingShowroom(), nil)
	storage.On("Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("new/banner.jpg", nil)
	repo.On("UpdateShowroomFields", mock.Anything, uint64(1), mock.Anything).Return(nil)

	banner := makeFileHeader("banner.jpg", 100)
	resp, err := svc.UpdateShowroom(context.Background(), uint64(1), ownerRoles(1), uint64(1),
		&showroom.UpdateShowroomRequest{}, nil, banner)
	assert.NoError(t, err)
	require.NotNil(t, resp.ShowroomBanner)
	assert.Equal(t, "new/banner.jpg", *resp.ShowroomBanner)
}

func TestUpdateShowroom_BannerUploadFails_BannerUnchanged(t *testing.T) {
	repo := new(mockShowroomRepo)
	storage := new(mockStorageProvider)
	svc := showroom.NewService(repo, storage, showroom.WithFileOpener(errorOpener(errors.New("open fail"))))

	existing := existingShowroom()
	repo.On("GetByID", mock.Anything, uint64(1)).Return(existing, nil)

	banner := makeFileHeader("banner.jpg", 100)
	resp, err := svc.UpdateShowroom(context.Background(), uint64(1), ownerRoles(1), uint64(1),
		&showroom.UpdateShowroomRequest{}, nil, banner)
	assert.NoError(t, err)
	assert.Equal(t, existing.ShowroomBanner, resp.ShowroomBanner)
	repo.AssertNotCalled(t, "UpdateShowroomFields")
}

func TestUpdateShowroom_UpdateShowroomFieldsError(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	repo.On("GetByID", mock.Anything, uint64(1)).Return(existingShowroom(), nil)
	repo.On("UpdateShowroomFields", mock.Anything, uint64(1), mock.Anything).Return(errors.New("db error"))

	_, err := svc.UpdateShowroom(context.Background(), uint64(1), ownerRoles(1), uint64(1),
		&showroom.UpdateShowroomRequest{Name: "New"}, nil, nil)
	assert.Error(t, err)
}

func TestUpdateShowroom_ManagerCanUpdate_Success(t *testing.T) {
	repo := new(mockShowroomRepo)
	svc := showroom.NewService(repo, new(mockStorageProvider))

	repo.On("GetByID", mock.Anything, uint64(1)).Return(existingShowroom(), nil)
	repo.On("UpdateShowroomFields", mock.Anything, uint64(1), mock.Anything).Return(nil)

	resp, err := svc.UpdateShowroom(context.Background(), uint64(7), managerRoles(1), uint64(1),
		&showroom.UpdateShowroomRequest{Name: "Updated By Manager", Geolocation: `{"city":"Mumbai"}`}, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, "Updated By Manager", resp.Name)
}
