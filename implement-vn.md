# implement.md: Triển khai mô hình Blockchain + Zero Trust bằng Ignite CLI và Cosmos SDK

> Đề tài 6: Kết hợp Blockchain và Zero Trust trong bảo vệ hệ thống mạng.
> Tài liệu này hướng dẫn xây dựng mô hình thử nghiệm (PoC) có thể chạy được, từ thiết kế, scaffold, cài đặt logic, gateway, kiểm thử đến đánh giá.
>
> Phiên bản công cụ có thể khác nhau. Hãy chạy `ignite version` và đối chiếu với tài liệu Ignite/Cosmos SDK tương ứng. Các đoạn mã Go bên dưới là **bản phác thảo** cần chỉnh theo tên kiểu (type) mà Ignite sinh ra trong project của bạn.

---

## 0. Nguyên tắc thiết kế (đọc trước)

Blockchain **không** tự bảo đảm danh tính hay tính đúng đắn của dữ liệu đầu vào. Vì vậy mô hình phân tách rõ:

| Thành phần | Vai trò | Nằm ở đâu |
|---|---|---|
| **PEP** (Policy Enforcement Point) | Chặn mọi yêu cầu truy cập, xác minh danh tính, thực thi quyết định | Off-chain (Gateway) |
| **PDP** (Policy Decision Point) | Đánh giá chính sách + ngữ cảnh, ra quyết định allow/deny | Off-chain (trong Gateway) đọc state từ chain |
| **Policy/Grant Store** | Nguồn sự thật có thể kiểm chứng về chính sách, cấp/thu hồi quyền | On-chain (module `ztaudit`) |
| **Audit Trail** | Nhật ký bất biến, có hash-chain, truy vết được | On-chain |
| **Identity Provider** | Xác thực người dùng (khóa, OIDC, mTLS) | Off-chain, có ràng buộc khóa công khai lên chain |

Các ràng buộc bắt buộc để chống "garbage in, garbage out":

1. **Mọi giao dịch đều được ký.** Danh tính on-chain = địa chỉ ký giao dịch, được kiểm tra bởi Cosmos SDK (không tự tin vào trường `subject` do client điền).
2. **Phân quyền ghi (authorization on write):** chỉ tài khoản `admin` mới cấp/thu hồi quyền hay đổi chính sách; chỉ tài khoản `gateway` đã đăng ký mới ghi được sự kiện truy cập.
3. **Quản lý khóa:** khóa admin và gateway tách biệt, lưu trong keyring (file/os/hardware), có quy trình xoay khóa (key rotation).
4. **Không lưu dữ liệu nhạy cảm trên chain.** Chỉ lưu định danh giả danh (pseudonymous ID) và hash của payload chi tiết.
5. **Giới hạn trust của gateway:** gateway là điểm tin cậy cho nhật ký truy cập. Giảm rủi ro bằng cách dùng nhiều gateway độc lập, hoặc ký kép (user ký challenge, gateway ký tx).

---

## 1. Kiến trúc tổng thể

```
 User/Device ──(1) challenge-response (ký bằng khóa riêng)──▶  PEP Gateway (Go)
                                                               │
                              (2) query grant/policy (gRPC)    │
                              ◀────────────────────────────────┤
                                  Cosmos chain (module ztaudit)│
                              (3) allow/deny + tx RecordAccess │
                              ─────────────────────────────────▶
                                                               │
                                              (4) proxy tới Protected Resource
 Admin CLI ──(tx: grant/revoke/update-policy, ký bằng khóa admin)──▶ Chain
 Auditor   ──(query audit trail + hash-chain verify)────────────▶ Chain
```

**Luồng chính**

1. Người dùng đăng ký danh tính (DID + public key) → on-chain `Identity`.
2. Admin cấp quyền `Grant{subject, resource, action, expires_at}` → on-chain, phát sự kiện.
3. Người dùng truy cập tài nguyên qua Gateway: Gateway gửi challenge, người dùng ký, Gateway xác minh chữ ký với public key lấy từ chain.
4. Gateway truy vấn `CheckAccess` (grant còn hiệu lực, chưa thu hồi, chưa hết hạn, identity active).
5. Gateway cho phép/từ chối và ghi `AccessRecord` (cả ALLOW lẫn DENY) vào chain.
6. Admin thu hồi quyền → `Grant.status = REVOKED`, các truy cập sau đó bị DENY.
7. Auditor truy vấn chuỗi audit, kiểm chứng hash-chain và (tùy chọn) Merkle proof.

---

## 2. Yêu cầu môi trường

- Go (phiên bản theo yêu cầu của Ignite/Cosmos SDK bạn dùng, thường ≥ 1.22)
- Ignite CLI: cài theo hướng dẫn chính thức, kiểm tra `ignite version`
- Docker (tùy chọn, để chạy nhiều node)
- `jq`, `curl`, `grpcurl` (kiểm thử)
- (tùy chọn) `k6` hoặc công cụ Go tự viết để đo tải

```bash
go version
ignite version
```

---

## 3. Khởi tạo chain và module

```bash
# Tạo chain, không tạo module mặc định
ignite scaffold chain zerotrust --address-prefix zt --no-module
cd zerotrust

# Tạo module chính chứa toàn bộ logic Zero Trust + audit
ignite scaffold module ztaudit
```

### 3.1. Tham số module (params)

```bash
ignite scaffold params admin:string gateway:string maxLogBatch:uint --module ztaudit
```

- `admin`: địa chỉ được phép cấp/thu hồi quyền và cập nhật chính sách.
- `gateway`: địa chỉ được phép ghi sự kiện truy cập (có thể mở rộng thành danh sách gateway).
- Đặt giá trị mặc định trong `genesis` (mục 8).

### 3.2. Các kiểu dữ liệu (state)

Dùng `--no-message` để **tự viết** message có kiểm tra quyền, tránh CRUD tự sinh cho phép ai cũng ghi được.

```bash
# Danh tính: did -> public key, vai trò, trạng thái
ignite scaffold map identity pubkey:string role:string status:string createdAt:int \
  --index did --module ztaudit --no-message

# Tài nguyên được bảo vệ
ignite scaffold map resource owner:string classification:string \
  --index resourceId --module ztaudit --no-message

# Quyền được cấp
ignite scaffold map grant subject:string resource:string action:string \
  grantor:string status:string expiresAt:int createdAt:int revokedAt:int \
  --index grantId --module ztaudit --no-message

# Chính sách: phiên bản, hash nội dung (nội dung đầy đủ để off-chain hoặc lưu gọn)
ignite scaffold map policy version:string contentHash:string updatedBy:string updatedAt:int \
  --index policyId --module ztaudit --no-message

# Nhật ký kiểm toán (list có ID tăng dần, dùng làm sequence cho hash-chain)
ignite scaffold list auditRecord eventType:string subject:string resource:string action:string \
  decision:string reason:string actor:string payloadHash:string prevHash:string \
  recordHash:string blockHeight:int timestamp:int \
  --module ztaudit --no-message
```

> Lưu ý: `ignite scaffold list` tạo bộ đếm (count) và ID tự tăng, thuận tiện cho chuỗi hash. Bạn sẽ tự viết hàm `AppendAudit` thay cho hàm CRUD sinh sẵn.

### 3.3. Message (giao dịch) và query

```bash
# Messages: tự viết logic kiểm tra quyền trong msg server
ignite scaffold message register-identity did pubkey role --module ztaudit --response ok:bool
ignite scaffold message register-resource resource-id classification --module ztaudit --response ok:bool
ignite scaffold message grant-access grant-id subject resource action expires-at:int --module ztaudit --response ok:bool
ignite scaffold message revoke-access grant-id reason --module ztaudit --response ok:bool
ignite scaffold message update-policy policy-id version content-hash --module ztaudit --response ok:bool
ignite scaffold message record-access subject resource action decision reason payload-hash --module ztaudit --response audit-id:uint

# Query kiểm tra quyền (PDP gọi qua gRPC) và xác minh chuỗi audit
ignite scaffold query check-access subject resource action --module ztaudit \
  --response allowed:bool,reason:string,grant-id:string
ignite scaffold query verify-chain from:uint to:uint --module ztaudit \
  --response valid:bool,broken-at:uint,last-hash:string
```

Sau đó kiểm tra build:

```bash
ignite chain build
```

---

## 4. Cài đặt logic nghiệp vụ (keeper / msg server)

### 4.1. Hàm kiểm tra quyền ghi

File `x/ztaudit/keeper/auth.go` (phác thảo):

```go
package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/types/errors"
)

func (k Keeper) mustBeAdmin(ctx sdk.Context, signer string) error {
	p := k.GetParams(ctx)
	if signer != p.Admin {
		return sdkerrors.Wrapf(errors.ErrUnauthorized, "signer %s is not admin", signer)
	}
	return nil
}

func (k Keeper) mustBeGateway(ctx sdk.Context, signer string) error {
	p := k.GetParams(ctx)
	if signer != p.Gateway {
		return sdkerrors.Wrapf(errors.ErrUnauthorized, "signer %s is not a registered gateway", signer)
	}
	return nil
}

var _ = fmt.Sprintf
```

> Với bản SDK mới, `GetParams` có thể trả `(Params, error)` và nhận `context.Context`. Hãy điều chỉnh theo code Ignite sinh ra.

### 4.2. Hash-chain cho nhật ký kiểm toán

Mỗi bản ghi tính:

```
recordHash = SHA256( id || eventType || subject || resource || action || decision ||
                     actor || payloadHash || blockHeight || timestamp || prevHash )
```

File `x/ztaudit/keeper/audit.go` (phác thảo):

```go
package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"zerotrust/x/ztaudit/types"
)

func computeHash(r types.AuditRecord) string {
	s := fmt.Sprintf("%d|%s|%s|%s|%s|%s|%s|%s|%d|%d|%s",
		r.Id, r.EventType, r.Subject, r.Resource, r.Action, r.Decision,
		r.Actor, r.PayloadHash, r.BlockHeight, r.Timestamp, r.PrevHash)
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// AppendAudit thêm một bản ghi vào chuỗi, tự gắn prevHash và recordHash.
func (k Keeper) AppendAudit(ctx sdk.Context, r types.AuditRecord) uint64 {
	// 1. Lấy bản ghi cuối để lấy prevHash
	count := k.GetAuditRecordCount(ctx)
	prev := "GENESIS"
	if count > 0 {
		last, found := k.GetAuditRecord(ctx, count-1)
		if found {
			prev = last.RecordHash
		}
	}
	// 2. Điền các trường do chain quyết định (KHÔNG tin client)
	r.BlockHeight = ctx.BlockHeight()
	r.Timestamp = ctx.BlockTime().Unix()
	r.PrevHash = prev
	r.Id = count
	r.RecordHash = computeHash(r)
	// 3. Lưu và tăng bộ đếm (dùng hàm AppendAuditRecord do Ignite sinh nếu có)
	return k.AppendAuditRecord(ctx, r)
}
```

Điểm quan trọng: `BlockHeight`, `Timestamp`, `PrevHash`, `Actor` do **chain** điền, client không thể giả mạo.

### 4.3. Msg server: cấp, thu hồi quyền, ghi truy cập

File `x/ztaudit/keeper/msg_server_grant_access.go` (phác thảo):

```go
func (k msgServer) GrantAccess(goCtx context.Context, msg *types.MsgGrantAccess) (*types.MsgGrantAccessResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// 1. Chỉ admin
	if err := k.mustBeAdmin(ctx, msg.Creator); err != nil {
		return nil, err
	}
	// 2. Identity và resource phải tồn tại và active
	id, ok := k.GetIdentity(ctx, msg.Subject)
	if !ok || id.Status != "ACTIVE" {
		return nil, sdkerrors.Wrap(types.ErrInvalidSubject, msg.Subject)
	}
	if _, ok := k.GetResource(ctx, msg.Resource); !ok {
		return nil, sdkerrors.Wrap(types.ErrUnknownResource, msg.Resource)
	}
	// 3. Không ghi đè grantId đã tồn tại
	if _, exists := k.GetGrant(ctx, msg.GrantId); exists {
		return nil, sdkerrors.Wrap(types.ErrGrantExists, msg.GrantId)
	}
	// 4. Hết hạn phải ở tương lai
	if msg.ExpiresAt <= ctx.BlockTime().Unix() {
		return nil, sdkerrors.Wrap(types.ErrInvalidExpiry, "expiresAt must be in the future")
	}

	k.SetGrant(ctx, types.Grant{
		GrantId:   msg.GrantId,
		Subject:   msg.Subject,
		Resource:  msg.Resource,
		Action:    msg.Action,
		Grantor:   msg.Creator,
		Status:    "ACTIVE",
		ExpiresAt: msg.ExpiresAt,
		CreatedAt: ctx.BlockTime().Unix(),
	})

	// 5. Ghi audit + phát event (cho indexer)
	k.AppendAudit(ctx, types.AuditRecord{
		EventType: "GRANT", Subject: msg.Subject, Resource: msg.Resource,
		Action: msg.Action, Decision: "N/A", Actor: msg.Creator,
		Reason: "grant " + msg.GrantId,
	})
	ctx.EventManager().EmitEvent(sdk.NewEvent("ztaudit_grant",
		sdk.NewAttribute("grant_id", msg.GrantId),
		sdk.NewAttribute("subject", msg.Subject),
		sdk.NewAttribute("resource", msg.Resource)))

	return &types.MsgGrantAccessResponse{Ok: true}, nil
}
```

`RevokeAccess` tương tự: chỉ admin, grant phải `ACTIVE`, đặt `Status = "REVOKED"`, `RevokedAt = blockTime`, ghi audit `REVOKE`.

`RecordAccess`: chỉ `gateway`; bắt buộc `Decision ∈ {ALLOW, DENY}`; ghi bản ghi `ACCESS` bằng `AppendAudit` và trả `audit-id`.

`RegisterIdentity`/`UpdatePolicy`: tương tự, kiểm tra quyền và ghi audit (`IDENTITY_REGISTER`, `POLICY_UPDATE`). Với `UpdatePolicy` ghi cả `contentHash` cũ và mới trong `payloadHash`/`reason` để chứng minh lịch sử thay đổi chính sách.

### 4.4. Query CheckAccess (PDP)

```go
func (k Keeper) CheckAccess(goCtx context.Context, req *types.QueryCheckAccessRequest) (*types.QueryCheckAccessResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	now := ctx.BlockTime().Unix()

	id, ok := k.GetIdentity(ctx, req.Subject)
	if !ok || id.Status != "ACTIVE" {
		return &types.QueryCheckAccessResponse{Allowed: false, Reason: "identity not active"}, nil
	}
	// Duyệt grant của subject (nên có index phụ subject->grantId để tối ưu)
	for _, g := range k.GetGrantsBySubject(ctx, req.Subject) {
		if g.Resource == req.Resource && g.Action == req.Action {
			if g.Status != "ACTIVE" {
				return &types.QueryCheckAccessResponse{Allowed: false, Reason: "grant revoked", GrantId: g.GrantId}, nil
			}
			if g.ExpiresAt <= now {
				return &types.QueryCheckAccessResponse{Allowed: false, Reason: "grant expired", GrantId: g.GrantId}, nil
			}
			return &types.QueryCheckAccessResponse{Allowed: true, Reason: "ok", GrantId: g.GrantId}, nil
		}
	}
	return &types.QueryCheckAccessResponse{Allowed: false, Reason: "no matching grant"}, nil
}
```

Tối ưu: tạo index phụ `subject|resource|action -> grantId` bằng prefix store riêng để tránh duyệt tuyến tính khi đo hiệu năng.

### 4.5. Query VerifyChain

Duyệt `from..to`, với mỗi bản ghi: (a) `recordHash == computeHash(record)`, (b) `record.PrevHash == prev.RecordHash`. Trả `valid=false` và `broken-at` nếu lệch. Auditor cũng nên tự kiểm chứng ở **phía client** để không phụ thuộc node.

---

## 5. Mô hình xác thực người dùng (off-chain) và Gateway

### 5.1. Ràng buộc khóa

- Mỗi người dùng/thiết bị có cặp khóa (secp256k1 hoặc ed25519). Public key đăng ký on-chain qua `register-identity` (do admin hoặc quy trình onboarding đã xác minh).
- Gateway **không** tin `subject` do client gửi. Subject suy ra từ chữ ký hợp lệ trên challenge.

### 5.2. Giao thức challenge-response

1. Client gọi `POST /auth/challenge {did}` → Gateway trả `nonce` ngẫu nhiên, TTL ngắn (ví dụ 30 giây), lưu trong bộ nhớ.
2. Client ký `H(nonce || did || resource || action || ts)` bằng khóa riêng, gọi `POST /access`.
3. Gateway lấy `pubkey` của `did` từ chain, xác minh chữ ký, đảm bảo `nonce` chỉ dùng một lần (chống replay).
4. Gateway gọi `CheckAccess` qua gRPC.
5. Gateway proxy yêu cầu nếu ALLOW; luôn gửi `MsgRecordAccess` (ALLOW hoặc DENY).

### 5.3. Cấu trúc project gateway

```
zerotrust/
├── app/ cmd/ x/ztaudit/ proto/ ...     # do Ignite sinh
└── gateway/
    ├── main.go
    ├── challenge.go     # nonce store + verify chữ ký
    ├── pdp.go           # gọi gRPC CheckAccess
    ├── recorder.go      # ký và gửi MsgRecordAccess (hàng đợi batch)
    └── config.yaml      # địa chỉ node, khóa gateway, TTL
```

### 5.4. Khung gateway (phác thảo)

```go
// gateway/main.go (rút gọn)
func handleAccess(w http.ResponseWriter, r *http.Request) {
	var req AccessReq // did, nonce, sig, resource, action, ts
	json.NewDecoder(r.Body).Decode(&req)

	// 1. Chống replay
	if !nonceStore.Consume(req.Nonce, req.DID) {
		deny(w, req, "invalid or reused nonce"); return
	}
	// 2. Lấy pubkey từ chain, xác minh chữ ký
	pub, err := chain.GetIdentityPubKey(req.DID)
	if err != nil || !verifySig(pub, req.SignedPayload(), req.Sig) {
		deny(w, req, "bad signature"); return
	}
	// 3. PDP
	res, _ := chain.CheckAccess(req.DID, req.Resource, req.Action)
	decision := "DENY"
	if res.Allowed { decision = "ALLOW" }

	// 4. Ghi audit (bất đồng bộ, theo batch) - hash payload chi tiết, không đưa PII lên chain
	recorder.Enqueue(MsgRecordAccess{
		Subject: req.DID, Resource: req.Resource, Action: req.Action,
		Decision: decision, Reason: res.Reason,
		PayloadHash: sha256Hex(req.RawBody),
	})

	if !res.Allowed { http.Error(w, res.Reason, 403); return }
	proxyToResource(w, r)
}
```

Chi tiết cần chú ý:
- `recorder` ký tx bằng **khóa gateway** riêng, quản lý `sequence` đúng để gửi tx liên tiếp (hoặc dùng nhiều tài khoản gateway).
- Chế độ ghi: **đồng bộ** (an toàn, chậm) hoặc **bất đồng bộ theo batch** (nhanh, có cửa sổ rủi ro mất log nếu gateway sập). Đo và so sánh cả hai ở mục 9.

---

## 6. Khởi chạy chuỗi thử nghiệm

```bash
ignite chain serve            # build, khởi tạo, chạy node dev với tài khoản mẫu alice, bob
```

Cấu hình tài khoản trong `config.yml`:

```yaml
version: 1
accounts:
  - name: admin
    coins: ["100000000stake"]
  - name: gateway
    coins: ["100000000stake"]
  - name: user1
    coins: ["1000000stake"]
  - name: user2
    coins: ["1000000stake"]
  - name: auditor
    coins: ["1000000stake"]
validators:
  - name: admin
    bonded: "100000000stake"
genesis:
  app_state:
    ztaudit:
      params:
        admin: "<địa chỉ admin>"
        gateway: "<địa chỉ gateway>"
        maxLogBatch: "100"
```

Lấy địa chỉ:

```bash
zerotrustd keys show admin -a --keyring-backend test
```

> Binary tên `zerotrustd` (tên chain + `d`). Với môi trường thật, dùng keyring `file` hoặc `os`, không dùng `test`.

---

## 7. Kịch bản thử nghiệm đầu-cuối (E2E)

Đặt biến:

```bash
ADMIN=$(zerotrustd keys show admin -a --keyring-backend test)
GW=$(zerotrustd keys show gateway -a --keyring-backend test)
U1=$(zerotrustd keys show user1 -a --keyring-backend test)
```

**Bước 1. Đăng ký danh tính và tài nguyên**

```bash
zerotrustd tx ztaudit register-identity "$U1" "<pubkey-base64>" employee \
  --from admin --keyring-backend test --chain-id zerotrust -y
zerotrustd tx ztaudit register-resource "db-finance" confidential \
  --from admin --keyring-backend test --chain-id zerotrust -y
```

**Bước 2. Cấp quyền**

```bash
EXP=$(( $(date +%s) + 3600 ))
zerotrustd tx ztaudit grant-access g-001 "$U1" db-finance read $EXP \
  --from admin --keyring-backend test --chain-id zerotrust -y
```

**Bước 3. Kiểm tra truy cập (kỳ vọng ALLOW)**

```bash
zerotrustd q ztaudit check-access "$U1" db-finance read
```

**Bước 4. Ghi sự kiện truy cập do gateway thực hiện**

```bash
zerotrustd tx ztaudit record-access "$U1" db-finance read ALLOW "ok" "<sha256-payload>" \
  --from gateway --keyring-backend test --chain-id zerotrust -y
```

**Bước 5. Thu hồi quyền, kiểm tra lại (kỳ vọng DENY)**

```bash
zerotrustd tx ztaudit revoke-access g-001 "employee left project" \
  --from admin --keyring-backend test --chain-id zerotrust -y
zerotrustd q ztaudit check-access "$U1" db-finance read
```

**Bước 6. Kiểm tra bảo mật (kỳ vọng thất bại)**

```bash
# user1 tự cấp quyền cho mình -> phải bị ErrUnauthorized
zerotrustd tx ztaudit grant-access g-evil "$U1" db-finance write $EXP \
  --from user1 --keyring-backend test --chain-id zerotrust -y
# user1 tự ghi log ALLOW giả -> phải bị ErrUnauthorized
zerotrustd tx ztaudit record-access "$U1" db-finance write ALLOW "fake" "x" \
  --from user1 --keyring-backend test --chain-id zerotrust -y
```

**Bước 7. Kiểm toán**

```bash
zerotrustd q ztaudit list-audit-record
zerotrustd q ztaudit verify-chain 0 100
```

---

## 8. Kiểm thử tự động (Go)

Đặt trong `x/ztaudit/keeper/*_test.go`. Các test bắt buộc:

| Nhóm | Test case |
|---|---|
| Phân quyền | Non-admin gọi GrantAccess/RevokeAccess/UpdatePolicy → lỗi Unauthorized |
| Phân quyền | Non-gateway gọi RecordAccess → lỗi Unauthorized |
| Logic quyền | Grant hợp lệ → CheckAccess allowed; hết hạn → denied; thu hồi → denied |
| Logic quyền | Grant cho identity không tồn tại/INACTIVE → lỗi |
| Toàn vẹn | Sau N thao tác, VerifyChain trả valid=true |
| Toàn vẹn | Sửa trực tiếp 1 bản ghi trong store (test) → VerifyChain báo `broken-at` đúng vị trí |
| Chống giả mạo | Trường `BlockHeight`, `Timestamp`, `PrevHash` luôn do chain điền, bỏ qua giá trị client |
| Idempotency | Grant trùng `grantId` → lỗi, không ghi đè |

Chạy:

```bash
go test ./x/ztaudit/... -v
```

Test gateway: replay nonce (phải bị từ chối), chữ ký sai, nonce hết hạn, subject giả.

---

## 9. Đánh giá (theo yêu cầu đề tài)

### 9.1. Khả năng kiểm toán

- Tập sự kiện: `IDENTITY_REGISTER`, `GRANT`, `REVOKE`, `POLICY_UPDATE`, `ACCESS(ALLOW/DENY)`.
- Chỉ số: **độ phủ sự kiện** (% sự kiện thực tế được ghi), **khả năng truy vết** (từ một truy cập suy ngược ra grant, grantor, policy version tại thời điểm đó), **thời gian truy vấn audit** theo `subject`/`resource`/khoảng thời gian.
- Thử nghiệm: mô phỏng sự cố "ai đã truy cập db-finance trước khi quyền bị thu hồi?" và đo thời gian trả lời.

### 9.2. Tính toàn vẹn

- Sửa/xóa/chèn bản ghi trong DB của một node → `verify-chain` và so sánh với node khác phát hiện.
- Đo **tỷ lệ phát hiện** và **vị trí phát hiện** (broken-at).
- Giới hạn cần nêu rõ: chain chứng minh log **không bị sửa sau khi ghi**, không chứng minh sự kiện đó **có thật** nếu gateway gian dối (xem mục 0).

### 9.3. Hiệu năng

Đo với các cấu hình: 1 node dev, 4 validator (Docker), ghi đồng bộ vs batch.

| Chỉ số | Cách đo |
|---|---|
| Độ trễ CheckAccess (query) | p50/p95/p99, 1000+ lần |
| Độ trễ end-to-end qua gateway (có/không ghi audit đồng bộ) | p50/p95/p99 |
| Thông lượng ghi (tx/s) | tăng dần tải đến khi mempool đầy / lỗi |
| Thời gian xác nhận | từ lúc gửi tx đến khi vào block |
| Ảnh hưởng của số lượng grant/bản ghi tới CheckAccess | 1k, 10k, 100k grant (so sánh có/không index phụ) |

Khung đo tải đơn giản (Go): chạy N goroutine gọi endpoint gateway, thu `time.Since` mỗi request, tính percentile. Có thể dùng `k6` cho HTTP.

### 9.3.1. Đường cơ sở (baseline)

Triển khai bản so sánh: **Zero Trust không blockchain** (gateway + PostgreSQL + audit log bình thường) để chứng minh chi phí/lợi ích tăng thêm của blockchain.

### 9.4. Chi phí vận hành

- **Lưu trữ:** kích thước 1 bản ghi audit (byte) × tốc độ sự kiện/ngày → dung lượng/tháng; đo tăng trưởng thư mục `~/.zerotrust/data`.
- **Tính toán:** CPU/RAM của node và gateway ở mức tải mục tiêu.
- **Gas/phí:** gas trung bình mỗi loại tx (xem `gas_used` trong kết quả tx); đặt `minimum-gas-prices` phù hợp.
- **Vận hành:** số node tối thiểu, sao lưu, nâng cấp, xoay khóa; so với baseline.
- Biện pháp giảm chi phí: ghi batch, chỉ đưa **hash** lên chain, lưu chi tiết off-chain (IPFS/DB) có tham chiếu bằng `payloadHash`, neo Merkle root theo chu kỳ thay vì từng sự kiện.

### 9.5. Bảng kết quả mẫu để điền

| Chỉ số | Baseline (DB) | Mô hình Blockchain | Ghi chú |
|---|---|---|---|
| Độ trễ CheckAccess p95 (ms) | | | |
| Độ trễ end-to-end p95 (ms) | | | |
| Thông lượng ghi (tx/s) | | | |
| Phát hiện sửa log (%) | | | |
| Dung lượng/1 triệu sự kiện (GB) | | | |
| Số node tối thiểu | | | |

---

## 10. Mô hình đe dọa (tóm tắt để đưa vào báo cáo)

| Mối đe dọa | Biện pháp trong mô hình | Rủi ro còn lại |
|---|---|---|
| Giả mạo danh tính | Challenge-response + pubkey đăng ký on-chain, chữ ký tx | Lộ khóa riêng → cần HSM/xoay khóa |
| Replay | Nonce dùng một lần + TTL | Đồng hồ lệch |
| Leo thang đặc quyền | Chỉ admin ghi grant/policy | Lộ khóa admin → dùng multisig / `x/authz` / `x/group` |
| Sửa log sau sự kiện | Hash-chain + đồng thuận nhiều validator | Đa số validator thông đồng |
| Gateway gian dối (không ghi hoặc ghi sai) | Nhiều gateway, đối chiếu log mạng độc lập | Là giới hạn cốt lõi: chain không biết sự thật ngoài chain |
| Lộ dữ liệu nhạy cảm | Chỉ lưu DID giả danh + hash | Phân tích liên kết (linkability) |
| Từ chối dịch vụ gateway/chain | Rate limit, hàng đợi, phí gas | Cần thiết kế HA |

Hướng nâng cao: dùng `x/authz`/`x/group` làm multisig cho admin, thêm điều kiện ngữ cảnh (thiết bị, vị trí, giờ) vào chính sách, và neo Merkle root lên chain công khai để chống thông đồng validator.

---

## 11. Lộ trình triển khai gợi ý (4 tuần)

| Tuần | Công việc | Kết quả |
|---|---|---|
| 1 | Scaffold chain/module, định nghĩa proto, params, genesis | Chain chạy được, CRUD khung |
| 2 | Msg server có kiểm tra quyền, hash-chain, CheckAccess, VerifyChain; unit test | Logic on-chain hoàn chỉnh + test pass |
| 3 | Gateway (challenge-response, PDP, recorder), kịch bản E2E | Demo cấp/thu hồi/truy cập có audit |
| 4 | Đo hiệu năng, chi phí, so sánh baseline, tấn công thử (tamper/replay), viết báo cáo | Số liệu + bảng đánh giá + kết luận |

---

## 12. Checklist hoàn thành

- [ ] Chain build và chạy bằng `ignite chain serve`
- [ ] Các message có kiểm tra quyền (admin/gateway), không có CRUD tự sinh mở
- [ ] Hash-chain đúng; `verify-chain` phát hiện được sửa đổi
- [ ] Gateway xác minh chữ ký, chống replay, ghi cả ALLOW và DENY
- [ ] Kịch bản cấp → truy cập → thu hồi → truy cập bị từ chối chạy được
- [ ] Unit test pass (`go test ./...`)
- [ ] Số liệu hiệu năng, chi phí và baseline đã thu thập
- [ ] Báo cáo nêu rõ giới hạn: blockchain không bảo đảm tính đúng của dữ liệu đầu vào

---

## 13. Tài liệu tham khảo

- NIST SP 800-207: Zero Trust Architecture
- Ignite CLI docs: https://docs.ignite.com
- Cosmos SDK docs: https://docs.cosmos.network
- W3C DID Core (tham khảo mô hình định danh phi tập trung)
