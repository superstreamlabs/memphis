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
import Modal from 'components/modal';
import Spinner from 'components/spinner';
import { ApiEndpoints } from 'const/apiEndpoints';
import { httpRequest } from 'services/http';
import { connectorTypesSource, connectorTypesSink } from 'connectors';

const ConnectorInfo = ({ open, clickOutside, connectorId }) => {
    const [loading, setLoading] = useState(false);
    const [info, setInfo] = useState(null);

    useEffect(() => {
        !open && setInfo(null);
        open && getConnectorInfo(connectorId);
    }, [open]);

    const getConnectorInfo = async (connectorId) => {
        setLoading(true);
        try {
            const data = await httpRequest('GET', `${ApiEndpoints.GET_CONNECTOR_DETAILS}?connector_id=${connectorId}`);
            arrangeData(data);
        } catch (error) {
        } finally {
            setLoading(false);
        }
    };

    const arrangeData = (data) => {
        let fieldInputs;
        if (data?.connector_type === 'source') {
            let field = connectorTypesSource.find((connector) => connector?.name?.toLocaleLowerCase() === data?.type);
            fieldInputs = field?.inputs?.Source;
        } else if (data?.connector_type === 'sink') {
            let field = connectorTypesSink.find((connector) => connector?.name?.toLocaleLowerCase() === data?.type);
            fieldInputs = field?.inputs?.Sink;
        }
        let formFields = [];
        fieldInputs?.forEach((field) => {
            formFields.push({ name: field?.display, value: data[field?.name] || data?.settings[field?.name] });
            if (field?.children) {
                field?.options?.forEach((option) => {
                    if (data?.settings[option?.toLocaleLowerCase()?.replace(/ /g, '_')]) {
                        formFields.push({ name: option, value: data?.settings[option?.toLocaleLowerCase()?.replace(/ /g, '_')] });
                    }
                });
            }
        });
        setInfo(formFields);
    };

    return (
        <Modal
            header={'Connector Information'}
            className={'modal-wrapper produce-modal'}
            width="550px"
            height="50vh"
            clickOutside={clickOutside}
            open={open}
            displayButtons={true}
            rBtnText={'Close'}
            rBtnClick={clickOutside}
        >
            <div className="connector-info">
                {loading && (
                    <div className="loader">
                        <Spinner />
                    </div>
                )}
                {!loading &&
                    info?.map((field, index) => {
                        return (
                            <div key={index} className="field-conainer">
                                <label className="field-name">{field?.name}</label>
                                <label className="field-value">{field?.value}</label>
                            </div>
                        );
                    })}
            </div>
        </Modal>
    );
};

export default ConnectorInfo;
