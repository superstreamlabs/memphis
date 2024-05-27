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
import { parsingDate } from 'services/valueConvertor';
import OverflowTip from 'components/tooltip/overflowtip';
const logsColumns = [
    {
        key: '1',
        title: 'Message',
        width: '300px'
    },
    {
        key: '2',
        title: 'Created at',
        width: '200px'
    }
];

const ConnectorError = ({ open, clickOutside, connectorId }) => {
    const [loading, setLoading] = useState(false);
    const [logs, setLogs] = useState(null);

    useEffect(() => {
        !open && setLogs(null);
        open && getConnectorErrors(connectorId);
    }, [open]);

    const getConnectorErrors = async (connectorId) => {
        setLoading(true);
        try {
            const data = await httpRequest('GET', `${ApiEndpoints.GET_CONNECTOR_ERRORS}?connector_id=${connectorId}`);
            setLogs(data?.logs);
        } catch (error) {
        } finally {
            setLoading(false);
        }
    };

    const purgeConnectorErrors = async () => {
        setLoading(true);
        try {
            await httpRequest('POST', ApiEndpoints.PURGE_CONNECTOR_ERRORS, {
                connector_id: connectorId
            });
            setLogs(null);
        } catch (error) {
        } finally {
            setLoading(false);
        }
    };

    return (
        <Modal
            header={"Connector's logs"}
            className={'modal-wrapper produce-modal'}
            width="550px"
            height="50vh"
            clickOutside={clickOutside}
            open={open}
            displayButtons={true}
            rBtnText={'Close'}
            lBtnText={'Purge logs'}
            rBtnClick={clickOutside}
            lBtnClick={purgeConnectorErrors}
        >
            <div className="connector-errors generic-list-wrapper">
                <div className="list">
                    <div className="coulmns-table">
                        {logsColumns?.map((column, index) => {
                            return (
                                <span key={index} style={{ width: column.width }}>
                                    {column.title}
                                </span>
                            );
                        })}
                    </div>
                    <div className="rows-wrapper">
                        {loading && (
                            <div className="loader">
                                <Spinner />
                            </div>
                        )}
                        {!loading && (!logs || logs?.length === 0) && <p className="no-logs">There are no logs to display</p>}
                        {logs?.map((row, index) => {
                            return (
                                <div className="pubSub-row" key={index}>
                                    <OverflowTip text={row?.message} width={'300px'}>
                                        {row?.message}
                                    </OverflowTip>
                                    <OverflowTip text={parsingDate(row?.created_at)} width={'200px'}>
                                        {parsingDate(row?.created_at)}
                                    </OverflowTip>
                                </div>
                            );
                        })}
                    </div>
                </div>
            </div>
        </Modal>
    );
};

export default ConnectorError;
