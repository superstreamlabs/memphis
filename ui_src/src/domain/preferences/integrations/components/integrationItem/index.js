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

import React, { useState, useEffect, useContext, useRef } from 'react';

import integrateIcon from '../../../../../assets/images/integrateIcon.svg';
import { capitalizeFirst } from '../../../../../services/valueConvertor';
import { Context } from '../../../../../hooks/store';
import Modal from '../../../../../components/modal';
import SlackIntegration from '../slackIntegration';
import S3Integration from '../s3Integration';
import Tag from '../../../../../components/tag';

const IntegrationItem = ({ value }) => {
    const [state, dispatch] = useContext(Context);
    const [modalIsOpen, modalFlip] = useState(false);
    const [integrateValue, setIntegrateValue] = useState({});

    const ref = useRef();
    ref.current = integrateValue;

    useEffect(() => {
        if (state.integrationsList?.length > 0) {
            checkIfUsed();
        }
    }, [state?.integrationsList]);

    const checkIfUsed = () => {
        let index = state.integrationsList?.findIndex((integration) => capitalizeFirst(integration.name) === value.name);
        setIntegrateValue(state.integrationsList[index]);
    };

    const modalContent = () => {
        switch (value.name) {
            case 'Slack':
                return (
                    <SlackIntegration
                        close={(data) => {
                            modalFlip(false);
                            setIntegrateValue(data);
                        }}
                        value={ref.current}
                    />
                );
            case 'Amazon S3':
                return (
                    <S3Integration
                        close={(data) => {
                            modalFlip(false);
                            setIntegrateValue(data);
                        }}
                        value={ref.current}
                    />
                );
            default:
                break;
        }
    };

    return (
        <>
            <integ-item is="3xd" onClick={() => (value.comingSoon ? null : modalFlip(true))}>
                {value.banner}
                {integrateValue && Object.keys(integrateValue)?.length !== 0 && (
                    <div className="integrate-icon">
                        <img src={integrateIcon} />
                    </div>
                )}
                <div className="integration-name">
                    {value.icon}
                    <div className="details">
                        <p>{value.name}</p>
                        <span>by {value.by}</span>
                    </div>
                </div>
                <p className="integration-description">{value.description} </p>
                <div className="category">
                    <Tag tag={value.category} />
                </div>
            </integ-item>
            <Modal className="integration-modal" height="95vh" width="720px" displayButtons={false} clickOutside={() => modalFlip(false)} open={modalIsOpen}>
                {modalContent()}
            </Modal>
        </>
    );
};

export default IntegrationItem;
