// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

import './style.scss';

import React, { useEffect, useState } from 'react';
import Button from 'components/button';
import Input from 'components/Input';

const DeleteItemsModal = ({ title, desc, handleDeleteSelected, buttontxt, textToConfirm, loader = false }) => {
    const [confirm, setConfirm] = useState('');

    useEffect(() => {
        const keyDownHandler = (event) => {
            if (event.key === 'Enter' && confirm === (textToConfirm || 'delete')) {
                handleDeleteSelected();
            }
        };
        document.addEventListener('keydown', keyDownHandler);
        return () => {
            document.removeEventListener('keydown', keyDownHandler);
        };
    }, [confirm]);

    return (
        <div className="delete-modal-wrapper">
            <p className="title">{title}</p>
            <p className="desc">{desc}</p>
            <div className="confirm-section">
                <p>
                    Please type <b>{textToConfirm || 'delete'}</b> to confirm.
                </p>
                <Input
                    placeholder={textToConfirm || 'delete'}
                    autoFocus={true}
                    type="text"
                    radiusType="semi-round"
                    colorType="black"
                    backgroundColorType="none"
                    borderColorType="gray-light"
                    height="48px"
                    onBlur={(e) => setConfirm(e.target.value)}
                    onChange={(e) => setConfirm(e.target.value)}
                    value={confirm}
                />
            </div>
            <div className="buttons">
                <Button
                    width="100%"
                    height="34px"
                    placeholder={buttontxt}
                    colorType="white"
                    radiusType="circle"
                    backgroundColorType="purple"
                    fontSize="12px"
                    fontFamily="InterSemiBold"
                    disabled={confirm !== (textToConfirm || 'delete') || loader}
                    isLoading={loader}
                    onClick={() => handleDeleteSelected()}
                />
            </div>
        </div>
    );
};

export default DeleteItemsModal;
