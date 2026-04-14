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

import React, { useState, useEffect } from 'react';
import { Modal, Checkbox } from '@douyinfe/semi-ui';

const EnableDisableUserModal = ({
  visible,
  onCancel,
  onConfirm,
  user,
  action,
  t,
}) => {
  const isDisable = action === 'disable';
  const [banInviter, setBanInviter] = useState(false);

  useEffect(() => {
    if (visible) {
      setBanInviter(false);
    }
  }, [visible]);

  const handleOk = () => {
    onConfirm({ banInviter });
  };

  return (
    <Modal
      title={isDisable ? t('确定要禁用此用户吗？') : t('确定要启用此用户吗？')}
      visible={visible}
      onCancel={onCancel}
      onOk={handleOk}
      type='warning'
    >
      <div>
        {isDisable ? t('此操作将禁用用户账户') : t('此操作将启用用户账户')}
      </div>
      {isDisable && user?.inviter_id > 0 && (
        <div style={{ marginTop: 12 }}>
          <Checkbox
            checked={banInviter}
            onChange={(e) => setBanInviter(e.target.checked)}
          >
            {t('同时封禁邀请人')} (ID: {user.inviter_id})
          </Checkbox>
        </div>
      )}
    </Modal>
  );
};

export default EnableDisableUserModal;
