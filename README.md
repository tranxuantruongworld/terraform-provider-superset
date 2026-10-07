# Terraform Provider for Apache Superset (`terraform-provider-superset`)

[![Go Version](https://img.shields.io/badge/go-1.27.1-blue.svg)](https://go.dev/)
[![Terraform Version](https://img.shields.io/badge/terraform-1.16.5-purple.svg)](https://www.terraform.io/)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

Dự án mã nguồn mở **Terraform Provider for Apache Superset** được xây dựng bằng **Go** và **Terraform Plugin Framework** của HashiCorp, cho phép quản lý phân quyền người dùng (RBAC - Role-Based Access Control) cho Apache Superset dưới dạng Infrastructure as Code (IaC).

---

## 🚀 Tính Năng Chính

- **Quản lý Role & Permissions (`superset_role`)**: Tạo và phân quyền chi tiết cho Role thông qua danh sách `permissions` (Permission View Menus - PVM).
- **Row-Level Security (`superset_row_level_security`)**: Quản lý các bộ lọc dữ liệu theo dòng (RLS) bằng câu lệnh SQL linh hoạt (`clause`), gắn với bảng (`tables`) và các nhóm quyền (`roles`).
- **Data Source Tra cứu Quyền (`data.superset_permission`)**: Tra cứu ID của các Permission View Menu theo cặp `(permission_name, view_menu_name)` như `can_read` trên `Dashboard`, `all_datasource_access`, v.v.
- **Quản lý User & Hỗ trợ SSO/UPN (`superset_user`)**: Hỗ trợ pre-provisioning tài khoản qua **UPN (User Principal Name)** cho người dùng đăng nhập qua Single Sign-On (Azure AD, GitLab, Vault, Keycloak). Chỉ cần khai báo `upn` và `roles`, hệ thống sẽ tự động cấu hình sẵn sàng.
- **Tự động xác thực REST API & CSRF**: Xác thực JWT token qua `/api/v1/security/login`, tự động lấy và truyền token `X-CSRFToken` cho mọi thao tác POST/PUT/DELETE.

---

## 📁 Cấu Trúc Dự Án

```text
terraform-provider-superset/
├── internal/
│   ├── client/                  # Superset REST API Client (Auth, Users, Roles, PVM, RLS, CSRF)
│   │   ├── client.go
│   │   ├── client_test.go       # Unit test giả lập HTTP server
│   │   ├── models.go            # Models dữ liệu JSON
│   │   ├── permissions.go       # API query Permission View Menus
│   │   ├── rls.go               # API CRUD Row-Level Security
│   │   ├── roles.go             # API CRUD Roles
│   │   └── users.go             # API CRUD Users
│   └── provider/                # Triển khai Terraform Plugin Framework
│       ├── provider.go           # Cấu hình Provider
│       ├── provider_test.go      # Kiểm thử Provider schema
│       ├── resource_role.go      # Resource superset_role (kèm permissions)
│       ├── resource_row_level_security.go # Resource superset_row_level_security
│       ├── resource_user.go      # Resource superset_user (hỗ trợ UPN SSO)
│       ├── data_source_permission.go # Data source superset_permission
│       ├── data_source_role.go   # Data source superset_role
│       └── data_source_user.go   # Data source superset_user
├── examples/
│   ├── dev.tfrc                 # Cấu hình dev_overrides cho Terraform CLI chạy local
│   └── main.tf                  # Ví dụ hoàn chỉnh về RBAC, RLS và UPN SSO
├── Makefile                     # Build, test, format, install
├── main.go                      # Điểm khởi chạy của Terraform Provider plugin
├── go.mod
├── go.sum
├── .gitignore
├── LICENSE                      # Apache 2.0
└── README.md
```

---

## 🛠 Yêu Cầu Cài Đặt (Prerequisites)

- **Go**: `1.27.1` (kiểm tra bằng `go version`)
- **Terraform**: `1.16.5` (kiểm tra bằng `terraform version`)

Cả 2 công cụ đã được cài đặt vào `~/.local/bin` (nằm sẵn trong `$PATH` hệ thống).

---

## 🔨 Hướng Dẫn Biên Dịch & Chạy Kiểm Thử

### 1. Biên dịch Provider
```bash
cd /home/xtruong/Desktop/open-source/terraform-provider-superset
make build
# hoặc: go build -o terraform-provider-superset .
```

### 2. Chạy Unit Tests
```bash
make test
# hoặc: go test -v ./...
```

---

## 🔐 Phân Tích SSO: Cách Superset Nhận Biết UPN Để Phân Quyền

Khi người dùng hoặc service đăng nhập vào Superset qua các dịch vụ Identity Provider (IdP) như **Azure AD / Microsoft Entra ID**, **GitLab**, **HashiCorp Vault**, hoặc **Keycloak**:

1. **Cơ chế hoạt động**:
   - Khi đăng nhập SSO qua OAuth2/OIDC, IdP trả về một chuỗi payload chứa các claims định danh người dùng:
     - **Azure AD (Entra ID)**: gửi thuộc tính `userPrincipalName` hoặc `upn` (ví dụ `nguyen.van.a@company.com`).
     - **GitLab**: gửi `preferred_username` hoặc `email`.
     - **Vault / Keycloak**: gửi `upn`, `preferred_username`, hoặc `email`.
   - Trong `superset_config.py`, lớp `CustomSecurityManager` của Flask-AppBuilder (FAB) sẽ ánh xạ thuộc tính này vào trường `username` của Superset.

2. **Cách kết hợp với Terraform**:
   - DevOps / Platform Admin khai báo user trước trong Terraform chỉ với **`upn`**:
     ```hcl
     resource "superset_user" "analyst" {
       upn   = "nguyen.van.a@company.com"
       roles = [superset_role.finance_analyst.id]
     }
     ```
   - Terraform Provider sẽ tự động:
     - Đặt `username = upn` (`nguyen.van.a@company.com`).
     - Đặt `email = upn`.
     - Tự sinh mật khẩu tạm thời ngẫu nhiên an toàn (vì user xác thực bằng SSO Token thay vì mật khẩu DB cục bộ).
     - Gán sẵn các Role và RLS tương ứng vào User.
   - Khi người dùng đăng nhập lần đầu tiên qua Azure/GitLab/Vault, Superset tìm thấy User có `username` trùng với `upn` và lập tức áp dụng ngay toàn bộ quyền mà bạn đã cấu hình qua Terraform mà không cần can thiệp thủ công!

3. **Mẫu cấu hình trong `superset_config.py`**:
```python
from superset.security import SupersetSecurityManager

class CustomSsoSecurityManager(SupersetSecurityManager):
    def oauth_user_info(self, provider, response=None):
        if provider == 'azure':
            data = self.appbuilder.sm.oauth_remotes[provider].get('me').json()
            # UPN là thuộc tính định danh chính xác nhất trong Azure AD
            upn = data.get('userPrincipalName') or data.get('upn') or data.get('mail')
            return {
                'username': upn,  # Khớp với 'upn' khai báo trong Terraform
                'email': data.get('mail') or upn,
                'first_name': data.get('givenName') or upn.split('@')[0],
                'last_name': data.get('surname') or 'SSO',
            }
        elif provider in ('gitlab', 'vault', 'keycloak'):
            data = self.appbuilder.sm.oauth_remotes[provider].get('user').json()
            upn = data.get('preferred_username') or data.get('email')
            return {
                'username': upn,
                'email': data.get('email') or upn,
                'first_name': data.get('name', 'SSO User'),
                'last_name': 'SSO',
            }
        return {}

CUSTOM_SECURITY_MANAGER = CustomSsoSecurityManager
```

---

## 💡 Ví Dụ Mẫu Cấu Hình HCL Đầy Đủ (`examples/main.tf`)

```hcl
terraform {
  required_providers {
    superset = {
      source = "xtruong/superset"
    }
  }
}

provider "superset" {
  base_url = "http://localhost:8088"
  username = "admin"
  password = "your_admin_password"
  insecure = true
}

# 1. Truy vấn các permission cụ thể
data "superset_permission" "read_dashboard" {
  permission_name = "can_read"
  view_menu_name   = "Dashboard"
}

data "superset_permission" "all_datasource" {
  permission_name = "all_datasource_access"
  view_menu_name   = "all_datasource_access"
}

# 2. Tạo Role và gán permissions chi tiết
resource "superset_role" "finance_analyst" {
  name = "Finance_Analyst"
  permissions = [
    data.superset_permission.read_dashboard.id,
    data.superset_permission.all_datasource.id,
  ]
}

# 3. Định nghĩa Row-Level Security (RLS)
resource "superset_row_level_security" "vn_finance_data" {
  name        = "VN_Finance_Only"
  filter_type = "Regular"
  clause      = "country = 'VN' AND department = 'finance'"
  tables      = [1] # Dataset ID
  roles       = [superset_role.finance_analyst.id]
}

# 4. Phân quyền User qua UPN (Single Sign-On)
resource "superset_user" "sso_analyst" {
  upn   = "nguyen.van.a@company.com"
  roles = [superset_role.finance_analyst.id]
}
```

---

## 📄 License
Dự án được phân phối dưới giấy phép [Apache License 2.0](LICENSE).
