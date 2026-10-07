# Terraform Provider for Apache Superset (`terraform-provider-superset`)

[![Go Version](https://img.shields.io/badge/go-1.27.1-blue.svg)](https://go.dev/)
[![Terraform Version](https://img.shields.io/badge/terraform-1.16.5-purple.svg)](https://www.terraform.io/)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

Dự án mã nguồn mở **Terraform Provider for Apache Superset** được xây dựng bằng **Go** và **Terraform Plugin Framework** của HashiCorp, cho phép quản lý phân quyền người dùng (RBAC - Role-Based Access Control) cho Apache Superset dưới dạng Infrastructure as Code (IaC).

---

## 🚀 Tính Năng Chính

- **Quản lý Role (`superset_role`)**: Tạo, cập nhật và xóa các vai trò tùy chỉnh (ví dụ: `Finance_Analyst`, `Data_Engineer`, `Marketing_Viewer`).
- **Quản lý User & Phân quyền (`superset_user`)**: Quản lý tài khoản người dùng, mật khẩu, trạng thái kích hoạt, và gán các role tương ứng.
- **Truy vấn Data Sources**:
  - `data.superset_role`: Truy vấn các role có sẵn của Superset (`Admin`, `Alpha`, `Gamma`, `Public`, `sql_lab`, v.v.) theo tên hoặc ID.
  - `data.superset_user`: Truy vấn thông tin người dùng theo `username` hoặc `id`.
- **Hỗ trợ REST API & CSRF**: Tự động xác thực JWT qua `/api/v1/security/login`, tự động lấy và gắn header `X-CSRFToken` khi gọi các mutation API.

---

## 📁 Cấu Trúc Dự Án

```text
terraform-provider-superset/
├── internal/
│   ├── client/              # Superset REST API Client (Auth, Users, Roles, CSRF)
│   │   ├── client.go
│   │   ├── client_test.go   # Unit test giả lập Superset server
│   │   ├── models.go        # Dữ liệu JSON models
│   │   ├── roles.go         # API methods cho Roles
│   │   └── users.go         # API methods cho Users
│   └── provider/            # Triển khai Terraform Plugin Framework
│       ├── provider.go       # Provider schema & configure
│       ├── provider_test.go  # Kiểm thử Provider schema
│       ├── resource_role.go  # Resource superset_role
│       ├── resource_user.go  # Resource superset_user
│       ├── data_source_role.go # Data source superset_role
│       └── data_source_user.go # Data source superset_user
├── examples/
│   ├── dev.tfrc             # Cấu hình dev_overrides cho Terraform CLI chạy local
│   └── main.tf              # Ví dụ mẫu triển khai phân quyền Superset
├── Makefile                 # Lệnh build, test, format, install
├── main.go                  # Điểm khởi chạy của Terraform Provider plugin
├── go.mod
├── go.sum
├── .gitignore
├── LICENSE                  # Apache 2.0
└── README.md
```

---

## 🛠 Yêu Cầu Cài Đặt (Prerequisites)

Dự án đã được thiết lập với phiên bản mới nhất:
- **Go**: `1.27.1` (kiểm tra bằng `go version`)
- **Terraform**: `1.16.5` (kiểm tra bằng `terraform version`)

Cả 2 công cụ đã được cài đặt vào `~/.local/bin` (đã nằm sẵn trong `$PATH` hệ thống).

---

## 🔨 Hướng Dẫn Biên Dịch & Chạy Kiểm Thử

### 1. Biên dịch Provider
```bash
cd /home/xtruong/Desktop/open-source/terraform-provider-superset
make build
# hoặc: go build -o terraform-provider-superset .
```
Lệnh trên sẽ sinh ra file nhị phân `terraform-provider-superset`.

### 2. Chạy Unit Tests
```bash
make test
# hoặc: go test -v ./...
```

---

## 📖 Hướng Dẫn Sử Dụng Local (Development Overrides)

Để kiểm thử provider trực tiếp với Terraform mà không cần xuất bản lên Terraform Registry, sử dụng tính năng `dev_overrides` có sẵn trong file `examples/dev.tfrc`.

### Bước 1: Khởi chạy / chuẩn bị Superset
Đảm bảo bạn có tài khoản Admin của Superset (mặc định username `admin`, password `admin`).

### Bước 2: Thiết lập biến môi trường
```bash
export TF_CLI_CONFIG_FILE="/home/xtruong/Desktop/open-source/terraform-provider-superset/examples/dev.tfrc"
export SUPERSET_BASE_URL="http://localhost:8088"
export SUPERSET_USERNAME="admin"
export SUPERSET_PASSWORD="your_admin_password"
```

### Bước 3: Áp dụng cấu hình Terraform
Vào thư mục `examples`:
```bash
cd examples
terraform plan
terraform apply
```

---

## 💡 Ví Dụ Mẫu Cấu Hình HCL (`examples/main.tf`)

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

# 1. Truy vấn các role có sẵn của Superset
data "superset_role" "gamma" {
  name = "Gamma"
}

# 2. Tạo role mới cho phòng ban
resource "superset_role" "finance_analyst" {
  name = "Finance_Analyst"
}

resource "superset_role" "data_engineer" {
  name = "Data_Engineer"
}

# 3. Tạo User và phân quyền (gán Roles)
resource "superset_user" "analyst_user" {
  username   = "john.analyst"
  first_name = "John"
  last_name  = "Doe"
  email      = "john.analyst@company.com"
  password   = "SuperSecret123!"
  active     = true

  roles = [
    data.superset_role.gamma.id,
    superset_role.finance_analyst.id,
  ]
}

resource "superset_user" "engineer_user" {
  username   = "alice.engineer"
  first_name = "Alice"
  last_name  = "Smith"
  email      = "alice.engineer@company.com"
  password   = "SuperSecret456!"
  active     = true

  roles = [
    superset_role.data_engineer.id,
  ]
}
```

---

## 📄 License
Dự án được phân phối dưới giấy phép [Apache License 2.0](LICENSE).
