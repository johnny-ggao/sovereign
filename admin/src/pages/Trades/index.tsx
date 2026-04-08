import { PageContainer, ProTable } from '@ant-design/pro-components';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { Button, Card, Col, Modal, Row, Statistic, Tag, Upload, message } from 'antd';
import type { UploadFile } from 'antd';
import React, { useEffect, useRef, useState } from 'react';
import { getTradeStats, getTrades, getTradeTemplateUrl, importTrades } from '@/services/api';
import dayjs from 'dayjs';

const pnlColor = (value: string): string =>
  parseFloat(value) >= 0 ? '#3f8600' : '#cf1322';

const formatPnl = (value: string): string =>
  parseFloat(value) >= 0 ? `+${value}` : value;

const tradeSourceTagMap: Record<string, { color: string; label: string }> = {
  api: { color: 'blue', label: 'API' },
  import: { color: 'green', label: '导入' },
  bot: { color: 'purple', label: '机器人' },
};

const TradesPage: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [stats, setStats] = useState<API.TradeStats | null>(null);
  const [statsLoading, setStatsLoading] = useState(true);
  const [importModalOpen, setImportModalOpen] = useState(false);
  const [importing, setImporting] = useState(false);
  const [fileList, setFileList] = useState<UploadFile[]>([]);
  const [importResult, setImportResult] = useState<API.TradeImportResult | null>(null);

  useEffect(() => {
    getTradeStats()
      .then((res) => {
        if (res.success && res.data) {
          setStats(res.data);
        }
      })
      .finally(() => setStatsLoading(false));
  }, []);

  const columns: ProColumns<API.TradeListItem>[] = [
    {
      title: '交易对',
      dataIndex: 'pair',
      copyable: true,
    },
    {
      title: '买入交易所',
      dataIndex: 'buy_exchange',
      hideInSearch: true,
      render: (_, r) => <Tag color="green">{r.buy_exchange}</Tag>,
    },
    {
      title: '卖出交易所',
      dataIndex: 'sell_exchange',
      hideInSearch: true,
      render: (_, r) => <Tag color="red">{r.sell_exchange}</Tag>,
    },
    {
      title: '买入价',
      dataIndex: 'buy_price',
      hideInSearch: true,
    },
    {
      title: '卖出价',
      dataIndex: 'sell_price',
      hideInSearch: true,
    },
    {
      title: '金额',
      dataIndex: 'amount',
      hideInSearch: true,
    },
    {
      title: '溢价率',
      dataIndex: 'premium_pct',
      hideInSearch: true,
      render: (_, r) => `${r.premium_pct}%`,
    },
    {
      title: '盈亏',
      dataIndex: 'pnl',
      hideInSearch: true,
      render: (_, r) => (
        <span style={{ color: pnlColor(r.pnl), fontWeight: 600 }}>
          {formatPnl(r.pnl)}
        </span>
      ),
    },
    {
      title: '手续费',
      dataIndex: 'fee',
      hideInSearch: true,
    },
    {
      title: '来源',
      dataIndex: 'source',
      hideInSearch: true,
      render: (_, record) => {
        const tag = tradeSourceTagMap[record.source] ?? { color: 'default', label: record.source };
        return <Tag color={tag.color}>{tag.label}</Tag>;
      },
    },
    {
      title: '执行时间',
      dataIndex: 'executed_at',
      hideInSearch: true,
      render: (_, r) => dayjs(r.executed_at).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '日期范围',
      dataIndex: 'dateRange',
      valueType: 'dateRange',
      hideInTable: true,
    },
  ];

  const resetImportState = () => {
    setImportModalOpen(false);
    setImporting(false);
    setFileList([]);
    setImportResult(null);
  };

  const handleImport = async () => {
    const selectedFile = fileList[0]?.originFileObj;
    if (!selectedFile) {
      message.error('请先选择Excel文件');
      return;
    }

    setImporting(true);
    try {
      const res = await importTrades(selectedFile as File);
      if (!res.success || !res.data) {
        message.error(res.error?.message ?? '导入交易失败');
        return;
      }
      setImportResult(res.data);
      if (res.data.errors.length > 0) {
        message.warning(`导入完成，成功 ${res.data.imported} 条，失败 ${res.data.errors.length} 条`);
      } else {
        message.success(`导入成功，共 ${res.data.imported} 条`);
      }
      actionRef.current?.reload();
    } catch (error: any) {
      message.error(error?.message ?? '导入交易失败');
    } finally {
      setImporting(false);
    }
  };

  return (
    <PageContainer>
      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={24} sm={8}>
          <Card loading={statsLoading}>
            <Statistic
              title="今日盈亏"
              value={stats?.pnl_1d ?? '0'}
              suffix="USDT"
              valueStyle={{ color: pnlColor(stats?.pnl_1d ?? '0') }}
            />
            <div style={{ marginTop: 8, fontSize: 13, color: '#999' }}>
              交易 {stats?.trade_count_1d ?? 0} 笔 · 用户利润{' '}
              {stats?.user_profit_1d ?? '0'} USDT
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card loading={statsLoading}>
            <Statistic
              title="7日盈亏"
              value={stats?.pnl_7d ?? '0'}
              suffix="USDT"
              valueStyle={{ color: pnlColor(stats?.pnl_7d ?? '0') }}
            />
            <div style={{ marginTop: 8, fontSize: 13, color: '#999' }}>
              交易 {stats?.trade_count_7d ?? 0} 笔 · 用户利润{' '}
              {stats?.user_profit_7d ?? '0'} USDT
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card loading={statsLoading}>
            <Statistic
              title="30日盈亏"
              value={stats?.pnl_30d ?? '0'}
              suffix="USDT"
              valueStyle={{ color: pnlColor(stats?.pnl_30d ?? '0') }}
            />
            <div style={{ marginTop: 8, fontSize: 13, color: '#999' }}>
              交易 {stats?.trade_count_30d ?? 0} 笔 · 用户利润{' '}
              {stats?.user_profit_30d ?? '0'} USDT
            </div>
          </Card>
        </Col>
      </Row>

      <ProTable<API.TradeListItem>
        headerTitle="套利交易历史"
        actionRef={actionRef}
        columns={columns}
        rowKey="id"
        search={{ labelWidth: 'auto' }}
        toolBarRender={() => [
          <Button
            key="template"
            onClick={() => {
              const token = localStorage.getItem('token');
              fetch(getTradeTemplateUrl(), {
                headers: { Authorization: `Bearer ${token}` },
              })
                .then((res) => res.blob())
                .then((blob) => {
                  const url = URL.createObjectURL(blob);
                  const a = document.createElement('a');
                  a.href = url;
                  a.download = 'trade_template.xlsx';
                  a.click();
                  URL.revokeObjectURL(url);
                })
                .catch(() => message.error('下载模板失败'));
            }}
          >
            下载模板
          </Button>,
          <Button key="import" type="primary" onClick={() => setImportModalOpen(true)}>
            导入交易
          </Button>,
        ]}
        request={async (params) => {
          const dateRange = params.dateRange;
          const res = await getTrades({
            page: params.current ?? 1,
            limit: params.pageSize ?? 20,
            pair: params.pair ?? '',
            date_from: dateRange?.[0] ?? '',
            date_to: dateRange?.[1] ?? '',
          });
          return {
            data: res.data ?? [],
            total: res.meta?.total ?? 0,
            success: res.success,
          };
        }}
        pagination={{ defaultPageSize: 20 }}
      />

      <Modal
        title="导入套利交易"
        open={importModalOpen}
        okText="开始导入"
        cancelText="取消"
        confirmLoading={importing}
        onOk={handleImport}
        onCancel={resetImportState}
      >
        <Upload.Dragger
          accept=".xlsx,.xls"
          maxCount={1}
          fileList={fileList}
          beforeUpload={(file) => {
            setFileList([file]);
            setImportResult(null);
            return false;
          }}
          onRemove={() => {
            setFileList([]);
            setImportResult(null);
          }}
        >
          <p>点击或拖拽Excel文件到此区域上传</p>
          <p>支持 `.xlsx` / `.xls` 格式，按模板列顺序导入套利交易记录。</p>
        </Upload.Dragger>

        {importResult && (
          <div style={{ marginTop: 16 }}>
            <div style={{ fontWeight: 600, marginBottom: 8 }}>
              导入结果: 成功 {importResult.imported} 条
            </div>
            <div style={{ color: '#666', marginBottom: 8 }}>
              错误数量: {importResult.errors.length}
            </div>
            {importResult.errors.length > 0 && (
              <div style={{ maxHeight: 160, overflowY: 'auto', padding: 12, background: '#fafafa', borderRadius: 6 }}>
                {importResult.errors.map((item) => (
                  <div key={item}>{item}</div>
                ))}
              </div>
            )}
          </div>
        )}
      </Modal>
    </PageContainer>
  );
};

export default TradesPage;
