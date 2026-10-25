package db

import (
	"fmt"
	"testing"
	"time"
)

// InsertSettingIfAbsent 是 auth.password_hash 它是一个安全网: 来电者的 GetSetting 检察官
// 数据库错误可能会导致瘫痪, 您可能会因同时请求而被打断。(bcrypt 有几十个
// 写入毫秒). 所以 "第一 1仅生鱼片设置" 的保证 if 它应该是一个主键约束，而不是一个判断。
// 做.
func TestInsertSettingIfAbsentDoesNotOverwrite(t *testing.T) {
	d, err := Open(testDSN(t))
	if err != nil {
		t.Skipf("postgres unavailable (%v) — skipping", err)
	}
	defer d.Close()

	// 使用仅供测试的密钥访问开发数据库中的实际数据。 auth.password_hash 请勿触摸.
	key := fmt.Sprintf("test.insert_if_absent.%d", time.Now().UnixNano())
	defer func() { _, _ = d.Exec(`DELETE FROM settings WHERE key=$1`, key) }()

	inserted, err := d.InsertSettingIfAbsent(key, "first")
	if err != nil {
		t.Fatal(err)
	}
	if !inserted {
		t.Fatal("第一次写作 inserted=true 一定是")
	}

	inserted, err = d.InsertSettingIfAbsent(key, "second")
	if err != nil {
		t.Fatal(err)
	}
	if inserted {
		t.Fatal("如果您已经有钥匙 inserted=false 一定是")
	}

	got, ok, err := d.GetSetting(key)
	if err != nil || !ok {
		t.Fatalf("GetSetting: ok=%v err=%v", ok, err)
	}
	if got != "first" {
		t.Fatalf("值 %q 覆盖着. %q 必须维持", got, "first")
	}
}
