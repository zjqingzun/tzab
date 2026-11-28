## 1. Overview

Client dùng để kiểm thử quy trình xác thực và phân quyền truy cập thông qua Zero Trust Gateway.

Các chức năng:
- Tạo cặp khóa Ed25519 (Public Key và Private Key).
- Ký request bằng Private Key.
- Gửi request đến Zero Trust Gateway.
- Xác thực Identity thông qua DID.
- Kiểm tra quyền truy cập thông qua Blockchain.

## 2. Project Structure

```text
client/
├── go.mod
├── main.go
├── models.go
├── private.key
└── public.key
```

Lưu ý:
- `private.key`: khóa bí mật được client sử dụng để ký request.
- `public.key`: khóa công khai được đăng ký lên Blockchain.
- Hai file key được tạo khi chạy client lần đầu và không được commit lên GitHub.

## 3. Prerequisites

- Go đã được cài đặt.
- Zero Trust Blockchain đã được khởi chạy.
- Zero Trust Gateway đã được khởi chạy.
- Có tài khoản Admin trên Blockchain.
- Có Resource đã được đăng ký trên Blockchain.

## 4. Start Blockchain

Configure Local Blockchain

Tạo file `config.local.yml` từ `config.yml`:

```bash
cp config.yml config.local.yml
```
Trong config.local.yml, khai báo các tài khoản sử dụng để kiểm thử:

```yaml
accounts:
  - name: admin
    mnemonic: "<ADMIN_MNEMONIC>"
    coins:
      - 20000token
      - 200000000stake

  - name: gateway
    mnemonic: "<GATEWAY_MNEMONIC>"
    coins:
      - 100000000stake

genesis:
  app_state:
    zt:
      params:
        admin: "<ADMIN_BLOCKCHAIN_ADDRESS>"
        gateway: "<GATEWAY_BLOCKCHAIN_ADDRESS>"
        maxLogBatch: "100"
```

Lưu ý bảo mật:
- Không commit config.local.yml lên Git.
- Không chia sẻ mnemonic hoặc Private Key.
- Địa chỉ trong genesis.app_state.zt.params phải khớp với tài khoản tương ứng.
Tại thư mục root của project:

```bash
ignite cosmos chain serve -c config.local.yml
```

## 5. Start Gateway

Mở terminal mới:

```bash
cd zerotrust/gateway
go run .
```

**Lưu ý:** Nếu chain đã được khởi tạo trước đó, thay đổi mnemonic trong file config không tự thay đổi tài khoản của genesis hiện tại. Cần bảo đảm mnemonic và địa chỉ đang sử dụng khớp với chain hiện tại.
Gateway mặc định chạy tại:

```text
http://localhost:8080
```

## 6. Generate Client Keys

Di chuyển vào thư mục client:

```bash
cd zerotrust/client
```

Chạy client lần đầu để tạo cặp khóa:

```bash
go run . \
  --gateway http://localhost:8080 \
  --did did:zt:user001 \
  --resource doc-001 \
  --action read \
  --method GET \
  --path /documents/doc-001
```

Client tạo hai file:

```text
private.key
public.key
```

**Lưu ý:** Giữ nguyên hai file key để sử dụng cho các lần kiểm thử tiếp theo. Không chia sẻ Private Key.

## 7. Register
### Identity

Lấy Public Key:

```bash
PUBKEY=$(cat public.key)
```

Đăng ký Identity lên Blockchain bằng tài khoản Admin:

```bash
zerotrustd tx zt register-identity \
  "did:zt:user001" \
  "$PUBKEY" \
  "employee" \
  --from admin \
  --chain-id zerotrust \
  --keyring-backend test \
  -y
```

Kiểm tra Identity:

```bash
zerotrustd query zt get-identity did:zt:user001
```

Identity cần có trạng thái `ACTIVE`.

### Resource

```bash
zerotrustd tx zt register-resource doc-001 confidential \
  --from admin \
  --chain-id zerotrust \
  --keyring-backend test \
  --gas auto \
  --gas-adjustment 1.5 \
  --yes
```

Kiểm tra resource
```bash
zerotrustd query zt get-resource doc-001
```

## 8. Grant Access

Admin cấp quyền đọc Resource `doc-001`:

```bash
zerotrustd tx zt grant-access grant-001 did:zt:user001 doc-001 read 0 \
  --from admin \
  --chain-id zerotrust \
  --keyring-backend test \
  --gas auto \
  --gas-adjustment 1.5 \
  --yes
```

Kiểm tra Grant:

```bash
zerotrustd query zt get-grant grant-001
```

Kiểm tra quyền truy cập:

```bash
zerotrustd query zt check-access did:zt:user001 doc-001 read
```

Kết quả mong đợi:

```text
allowed: true
```

## 9. Send Access Request

Chạy lại client để gửi request đã ký:

```bash
go run . \
  --gateway http://localhost:8080 \
  --did did:zt:user001 \
  --resource doc-001 \
  --action read \
  --method GET \
  --path /documents/doc-001
```

Gateway thực hiện:
1. Xác thực chữ ký request.
2. Kiểm tra Identity.
3. Kiểm tra quyền truy cập trên Blockchain.
4. Ghi nhận Audit Record.
5. Cho phép hoặc từ chối request.
6. Forward request đến Resource Server nếu được phép.

## 10. Revoke Access

Admin thu hồi quyền:

```bash
zerotrustd tx zt revoke-access grant-001 "manual revocation" \
  --from admin \
  --chain-id zerotrust \
  --keyring-backend test \
  --gas auto \
  --gas-adjustment 1.5 \
  --yes
```

Kiểm tra lại quyền:

```bash
zerotrustd query zt check-access did:zt:user001 doc-001 read
```

Kết quả mong đợi:

```text
allowed: false
``` 

## Kiểm tra các lần access

```bash
zerotrustd query zt check-access did:zt:user001 doc-001 read
```