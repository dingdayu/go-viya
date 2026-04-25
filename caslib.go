package viya

import (
	"context"
	"fmt"
)

// CAS_SERVER_NAME 是默认的 CAS 服务器名称
// 在实际使用中，只需要 UnLoadCasLibTableInMemory 对应的 表，当 Viya 可视化报表加载数据时，就会自动重载数据到内存中了
// caslib 操作需要在 Password 认证模式下才能操作成功

const CAS_SERVER_NAME = "cas-shared-default"

// LoadCasLibTableToMemory 将某个 CAS 库的表加载到内存中
func (c *Client) LoadCasLibTableToMemory(ctx context.Context, casLibName, table string, replace bool, scope string) error {
	body := map[string]any{
		"outputCaslibName": casLibName,
		"outputTableName":  table,
		"replace":          replace,
		"scope":            scope,
	}

	resp, err := c.client.R().SetContext(ctx).
		SetQueryParam("value", "loaded").
		SetBody(body).
		Put(fmt.Sprintf("/casManagement/servers/%s/caslibs/%s/tables/%s/state", CAS_SERVER_NAME, casLibName, table))
	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("failed to load CAS library table: %s", resp.Status())
	}
	return nil
}

// UnLoadCasLibTableInMemory 从内存中移除某个 CAS 库的表
func (c *Client) UnLoadCasLibTableInMemory(ctx context.Context, casLibName, table string) error {
	resp, err := c.client.R().SetContext(ctx).
		SetQueryParam("value", "unloaded").
		Put(fmt.Sprintf("/casManagement/servers/%s/caslibs/%s/tables/%s/state", CAS_SERVER_NAME, casLibName, table))
	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("failed to load CAS library table: %s, %s", resp.Status(), resp.String())
	}
	return nil
}
