// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis-broker/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

import './style.scss';

import React, { useEffect, useState } from 'react';
import Button from '../../../../../components/button';
import CheckboxComponent from '../../../../../components/checkBox';
import Input from '../../../../../components/Input';

const PurgeStationModal = ({ title, desc, handlePurgeSelected, cancel, loader = false, msgsDisabled = false, dlsDisabled = false }) => {
    const [confirm, setConfirm] = useState('');
    const [purgeData, setPurgeData] = useState({
        purge_station: false,
        purge_dls: false
    });

    useEffect(() => {
        const keyDownHandler = (event) => {
            if (event.key === 'Enter' && confirm === 'purge') {
                handlePurgeSelected();
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
            <div className="checkbox-body">
                <span>
                    <CheckboxComponent
                        checked={purgeData.purge_station}
                        id={'purge_station'}
                        onChange={(e) => setPurgeData({ ...purgeData, purge_station: !purgeData.purge_station })}
                        disabled={msgsDisabled}
                        name={'purge_station'}
                    />
                    <p>Messages</p>
                </span>
                <span>
                    <CheckboxComponent
                        checked={purgeData.purge_dls}
                        id={'purge_dls'}
                        onChange={(e) => setPurgeData({ ...purgeData, purge_dls: !purgeData.purge_dls })}
                        disabled={dlsDisabled}
                        name={'purge_dls'}
                    />
                    <p>Dead-letter</p>
                </span>
            </div>
            <div className="confirm-section">
                <p>
                    Please type <b> purge</b> to confirm.
                </p>

                <Input
                    placeholder="purge"
                    autoFocus={true}
                    type="text"
                    radiusType="semi-round"
                    colorType="black"
                    backgroundColorType="none"
                    borderColorType="gray-light"
                    height="43px"
                    onBlur={(e) => setConfirm(e.target.value)}
                    onChange={(e) => setConfirm(e.target.value)}
                    value={confirm}
                />
            </div>
            <div className="buttons">
                <Button
                    width="200px"
                    height="34px"
                    placeholder="Close"
                    colorType="navy"
                    radiusType="circle"
                    border="gray"
                    backgroundColorType="none"
                    fontSize="12px"
                    fontFamily="InterSemiBold"
                    onClick={cancel}
                />
                <Button
                    width="200px"
                    height="34px"
                    placeholder="Confirm"
                    colorType="white"
                    radiusType="circle"
                    backgroundColorType="purple"
                    fontSize="12px"
                    fontFamily="InterSemiBold"
                    disabled={confirm !== 'purge' || (!purgeData.purge_dls && !purgeData.purge_station)}
                    isLoading={loader}
                    onClick={() => handlePurgeSelected(purgeData)}
                />
            </div>
        </div>
    );
};

export default PurgeStationModal;
