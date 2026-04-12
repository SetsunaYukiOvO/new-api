import React, { useState } from 'react';
import { Button, Input, Modal, Typography, Banner, Steps } from '@douyinfe/semi-ui';
import { IconCopy, IconTick } from '@douyinfe/semi-icons';
import { API, showError, showSuccess, copy } from '../../../../helpers';

const { Text, Paragraph } = Typography;

const QQBindModal = ({ t, visible, onClose, onSuccess }) => {
  const [step, setStep] = useState(0); // 0: enter QQ, 1: show code, 2: verifying
  const [qqNumber, setQqNumber] = useState('');
  const [verifyCode, setVerifyCode] = useState('');
  const [loading, setLoading] = useState(false);
  const [countdown, setCountdown] = useState(0);

  const resetState = () => {
    setStep(0);
    setQqNumber('');
    setVerifyCode('');
    setLoading(false);
    setCountdown(0);
  };

  const handleClose = () => {
    resetState();
    onClose();
  };

  const generateCode = async () => {
    if (!qqNumber || !/^\d{5,12}$/.test(qqNumber)) {
      showError(t('请输入正确的QQ号（5-12位数字）'));
      return;
    }
    setLoading(true);
    try {
      const res = await API.post('/api/user/qq_bind/code', {
        qq_number: qqNumber,
      });
      const { success, message, data } = res.data;
      if (success) {
        setVerifyCode(data.code);
        setStep(1);
        // Start countdown
        setCountdown(300);
        const timer = setInterval(() => {
          setCountdown((prev) => {
            if (prev <= 1) {
              clearInterval(timer);
              return 0;
            }
            return prev - 1;
          });
        }, 1000);
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.response?.data?.message || t('请求失败'));
    } finally {
      setLoading(false);
    }
  };

  const verifyBind = async () => {
    setLoading(true);
    try {
      const res = await API.post('/api/user/qq_bind/verify', {
        qq_number: qqNumber,
        code: verifyCode,
      });
      const { success, message } = res.data;
      if (success) {
        showSuccess(message);
        handleClose();
        if (onSuccess) onSuccess();
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.response?.data?.message || t('验证失败'));
    } finally {
      setLoading(false);
    }
  };

  const formatCountdown = (seconds) => {
    const m = Math.floor(seconds / 60);
    const s = seconds % 60;
    return `${m}:${s.toString().padStart(2, '0')}`;
  };

  return (
    <Modal
      title={t('绑定QQ')}
      visible={visible}
      onCancel={handleClose}
      footer={null}
      size='small'
      centered
      className='modern-modal'
    >
      <div className='space-y-4 py-4'>
        {step === 0 && (
          <div className='space-y-4'>
            <div>
              <Text strong>{t('请输入你的QQ号')}</Text>
              <Input
                className='mt-2'
                placeholder={t('请输入QQ号')}
                value={qqNumber}
                onChange={setQqNumber}
                size='large'
              />
            </div>
            <Button
              type='primary'
              theme='solid'
              block
              loading={loading}
              onClick={generateCode}
            >
              {t('获取验证码')}
            </Button>
          </div>
        )}

        {step === 1 && (
          <div className='space-y-4'>
            <Banner
              type='info'
              description={t('请将你的QQ昵称修改为以下验证码，修改后点击验证按钮')}
              closeIcon={null}
            />

            <div className='text-center p-4 bg-slate-50 dark:bg-slate-800 rounded-lg'>
              <Text type='secondary' size='small'>
                {t('验证码')}
              </Text>
              <Paragraph
                className='mt-2'
                copyable={{ content: verifyCode }}
                style={{ fontSize: 20, fontWeight: 'bold', fontFamily: 'monospace' }}
              >
                {verifyCode}
              </Paragraph>
            </div>

            <div className='text-center'>
              <Text type='tertiary' size='small'>
                {t('QQ号')}: {qqNumber}
                {countdown > 0 && (
                  <span className='ml-4'>
                    {t('剩余时间')}: {formatCountdown(countdown)}
                  </span>
                )}
              </Text>
            </div>

            <div className='space-y-3'>
              <Banner
                type='warning'
                description={
                  <div>
                    <div>{t('操作步骤：')}</div>
                    <div>{t('1. 复制上方验证码')}</div>
                    <div>{t('2. 打开QQ → 点击头像 → 编辑资料 → 昵称')}</div>
                    <div>{t('3. 将QQ昵称修改为验证码并保存')}</div>
                    <div>{t('4. 回到此页面点击下方验证按钮')}</div>
                  </div>
                }
                closeIcon={null}
              />
            </div>

            <div className='flex gap-2'>
              <Button
                block
                onClick={() => setStep(0)}
              >
                {t('返回')}
              </Button>
              <Button
                type='primary'
                theme='solid'
                block
                loading={loading}
                disabled={countdown === 0}
                onClick={verifyBind}
              >
                {countdown === 0 ? t('验证码已过期') : t('我已修改昵称，验证绑定')}
              </Button>
            </div>
          </div>
        )}
      </div>
    </Modal>
  );
};

export default QQBindModal;
