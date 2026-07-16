/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React, { useEffect, useState, useRef } from 'react';
import { Banner, Button, Form, Row, Col, Spin } from '@douyinfe/semi-ui';
import {
  API,
  removeTrailingSlash,
  showError,
  showSuccess,
} from '../../../helpers';
import { useTranslation } from 'react-i18next';
import { BookOpen } from 'lucide-react';

export default function SettingsPaymentGatewayBepusdt(props) {
  const { t } = useTranslation();
  const sectionTitle = props.hideSectionTitle ? undefined : t('BEPUSDT 设置');
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    BepusdtEnabled: false,
    BepusdtUrl: '',
    BepusdtApiKey: '',
    BepusdtFiat: 'USD',
    BepusdtCurrencies: '',
    BepusdtTradeType: '',
    BepusdtUnitPrice: 1.0,
    BepusdtMinTopUp: 1,
    BepusdtNotifyUrl: '',
    BepusdtReturnUrl: '',
  });
  const [originInputs, setOriginInputs] = useState({});
  const formApiRef = useRef(null);

  useEffect(() => {
    if (props.options && formApiRef.current) {
      const currentInputs = {
        BepusdtEnabled:
          props.options.BepusdtEnabled === true ||
          props.options.BepusdtEnabled === 'true',
        BepusdtUrl: props.options.BepusdtUrl || '',
        BepusdtApiKey: props.options.BepusdtApiKey || '',
        BepusdtFiat: props.options.BepusdtFiat || 'USD',
        BepusdtCurrencies: props.options.BepusdtCurrencies || '',
        BepusdtTradeType: props.options.BepusdtTradeType || '',
        BepusdtUnitPrice:
          props.options.BepusdtUnitPrice !== undefined
            ? parseFloat(props.options.BepusdtUnitPrice)
            : 1.0,
        BepusdtMinTopUp:
          props.options.BepusdtMinTopUp !== undefined
            ? parseInt(props.options.BepusdtMinTopUp)
            : 1,
        BepusdtNotifyUrl: props.options.BepusdtNotifyUrl || '',
        BepusdtReturnUrl: props.options.BepusdtReturnUrl || '',
      };
      setInputs(currentInputs);
      setOriginInputs({ ...currentInputs });
      formApiRef.current.setValues(currentInputs);
    }
  }, [props.options]);

  const handleFormChange = (values) => {
    setInputs(values);
  };

  const submitBepusdtSetting = async () => {
    setLoading(true);
    try {
      const options = [];

      if (originInputs.BepusdtEnabled !== inputs.BepusdtEnabled) {
        options.push({
          key: 'BepusdtEnabled',
          value: inputs.BepusdtEnabled ? 'true' : 'false',
        });
      }
      if (inputs.BepusdtUrl !== originInputs.BepusdtUrl) {
        options.push({
          key: 'BepusdtUrl',
          value: removeTrailingSlash(inputs.BepusdtUrl || ''),
        });
      }
      if (inputs.BepusdtApiKey && inputs.BepusdtApiKey !== '') {
        options.push({ key: 'BepusdtApiKey', value: inputs.BepusdtApiKey });
      }
      if (inputs.BepusdtFiat !== originInputs.BepusdtFiat) {
        options.push({
          key: 'BepusdtFiat',
          value: inputs.BepusdtFiat || 'USD',
        });
      }
      if (inputs.BepusdtCurrencies !== originInputs.BepusdtCurrencies) {
        options.push({
          key: 'BepusdtCurrencies',
          value: inputs.BepusdtCurrencies || '',
        });
      }
      if (inputs.BepusdtTradeType !== originInputs.BepusdtTradeType) {
        options.push({
          key: 'BepusdtTradeType',
          value: inputs.BepusdtTradeType || '',
        });
      }
      if (
        inputs.BepusdtUnitPrice !== undefined &&
        inputs.BepusdtUnitPrice !== null &&
        inputs.BepusdtUnitPrice !== originInputs.BepusdtUnitPrice
      ) {
        options.push({
          key: 'BepusdtUnitPrice',
          value: inputs.BepusdtUnitPrice.toString(),
        });
      }
      if (
        inputs.BepusdtMinTopUp !== undefined &&
        inputs.BepusdtMinTopUp !== null &&
        inputs.BepusdtMinTopUp !== originInputs.BepusdtMinTopUp
      ) {
        options.push({
          key: 'BepusdtMinTopUp',
          value: inputs.BepusdtMinTopUp.toString(),
        });
      }
      if (inputs.BepusdtNotifyUrl !== originInputs.BepusdtNotifyUrl) {
        options.push({
          key: 'BepusdtNotifyUrl',
          value: inputs.BepusdtNotifyUrl || '',
        });
      }
      if (inputs.BepusdtReturnUrl !== originInputs.BepusdtReturnUrl) {
        options.push({
          key: 'BepusdtReturnUrl',
          value: inputs.BepusdtReturnUrl || '',
        });
      }

      if (options.length === 0) {
        showSuccess(t('未更改任何配置'));
        setLoading(false);
        return;
      }

      const requestQueue = options.map((opt) =>
        API.put('/api/option/', {
          key: opt.key,
          value: opt.value,
        }),
      );

      const results = await Promise.all(requestQueue);
      const errorResults = results.filter((res) => !res.data.success);
      if (errorResults.length > 0) {
        errorResults.forEach((res) => {
          showError(res.data.message);
        });
      } else {
        showSuccess(t('更新成功'));
        setOriginInputs({ ...inputs });
        props.refresh?.();
      }
    } catch (error) {
      showError(t('更新失败'));
    }
    setLoading(false);
  };

  return (
    <Spin spinning={loading}>
      <Form
        initValues={inputs}
        onValueChange={handleFormChange}
        getFormApi={(api) => (formApiRef.current = api)}
      >
        <Form.Section text={sectionTitle}>
          <Banner
            type='info'
            icon={<BookOpen size={16} />}
            description={
              <>
                {t('回调地址')}：
                {props.options.ServerAddress
                  ? removeTrailingSlash(props.options.ServerAddress)
                  : t('网站地址')}
                /api/bepusdt/notify
                <br />
                {t(
                  'Trade Type 为空时使用 create-order；填写后使用 create-transaction。',
                )}
              </>
            }
            style={{ marginBottom: 16 }}
          />
          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.Switch
                field='BepusdtEnabled'
                label={t('启用 BEPUSDT')}
              />
            </Col>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.Input
                field='BepusdtUrl'
                label={t('网关地址')}
                placeholder='https://pay.example.com'
              />
            </Col>
          </Row>
          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.Input
                field='BepusdtApiKey'
                label={t('API Key')}
                mode='password'
                placeholder={t('留空则不更新')}
              />
            </Col>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.Input
                field='BepusdtFiat'
                label={t('法币')}
                placeholder='USD'
              />
            </Col>
          </Row>
          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.Input
                field='BepusdtCurrencies'
                label={t('加密货币（可选）')}
                placeholder='USDT.TRC20,USDT.BEP20'
              />
            </Col>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.Input
                field='BepusdtTradeType'
                label={t('Trade Type（可选）')}
                placeholder='usdt'
              />
            </Col>
          </Row>
          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.InputNumber
                field='BepusdtUnitPrice'
                label={t('单价')}
                min={0}
                step={0.01}
              />
            </Col>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.InputNumber
                field='BepusdtMinTopUp'
                label={t('最低充值数量')}
                min={1}
                step={1}
              />
            </Col>
          </Row>
          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.Input
                field='BepusdtNotifyUrl'
                label={t('回调地址覆盖（可选）')}
                placeholder='https://api.example.com/api/bepusdt/notify'
              />
            </Col>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.Input
                field='BepusdtReturnUrl'
                label={t('返回地址覆盖（可选）')}
                placeholder='https://example.com/console/topup'
              />
            </Col>
          </Row>
          <Button onClick={submitBepusdtSetting} style={{ marginTop: 16 }}>
            {t('保存 BEPUSDT 设置')}
          </Button>
        </Form.Section>
      </Form>
    </Spin>
  );
}
