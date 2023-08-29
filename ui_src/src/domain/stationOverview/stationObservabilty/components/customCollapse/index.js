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
import { Col, Collapse } from 'antd';

import { ReactComponent as CollapseArrowIcon } from '../../../../../assets/images/collapseArrow.svg';
import { ReactComponent as WarningCircleIcon } from '../../../../../assets/images/warningCircle.svg';
import OverflowTip from '../../../../../components/tooltip/overflowtip';
import Copy from '../../../../../components/copy';
import { messageParser } from '../../../../../services/valueConvertor';
import SegmentButton from '../../../../../components/segmentButton';
import TooltipComponent from '../../../../../components/tooltip/tooltip';
import { LOCAL_STORAGE_MSG_PARSER } from '../../../../../const/localStorageConsts';
import ConsumerWithStatus from '../../../../../components/consumerWithStatus';

const { Panel } = Collapse;

const CustomCollapse = ({ status, data, header, defaultOpen, collapsible, message, tooltip, schemaType, consumerList = false }) => {
    const [activeKey, setActiveKey] = useState(defaultOpen ? ['1'] : []);
    const [parser, setParser] = useState(schemaType || localStorage.getItem(LOCAL_STORAGE_MSG_PARSER) || 'string');
    const [payload, setPayload] = useState(data);

    useEffect(() => {
        if (header === 'Payload') {
            setPayload(messageParser(parser, data));
        }
    }, [parser, data]);

    const onChange = (key) => {
        setActiveKey(key);
    };

    const drawHeaders = (headers) => {
        let obj = [];
        for (const property in headers) {
            obj.push(
                <div className="headers-container" key={property}>
                    <p>{property}</p>
                    <div className="copy-section">
                        <Copy data={headers[property]}></Copy>
                        <OverflowTip text={headers[property]} width={'calc(100% - 10px)'}>
                            {headers[property]}
                        </OverflowTip>
                    </div>
                </div>
            );
        }
        return obj;
    };

    return (
        <Collapse ghost defaultActiveKey={activeKey} onChange={onChange} className="custom-collapse">
            <Panel
                showArrow={false}
                collapsible={collapsible || data?.length === 0 || (data !== undefined && Object?.keys(data)?.length === 0) ? 'disabled' : null}
                className={header === 'Payload' ? 'payload-header' : consumerList ? 'consumer-list-header' : ''}
                header={
                    <TooltipComponent text={tooltip}>
                        <div className="collapse-header">
                            <div className="first-row">
                                <p className="title">
                                    {header}
                                    {header === 'Headers' && <span className="consumer-number">{data !== undefined ? Object?.keys(data)?.length : ''}</span>}
                                    {header === 'Validation error' && <WarningCircleIcon className="validation-image" />}
                                </p>
                                <status is="x3d">
                                    {activeKey[0] === '1' ? (
                                        <CollapseArrowIcon className="collapse-arrow-icon open" />
                                    ) : (
                                        <CollapseArrowIcon className="collapse-arrow-icon close" />
                                    )}
                                </status>
                            </div>
                        </div>
                    </TooltipComponent>
                }
                key="1"
            >
                {consumerList &&
                    data?.length > 0 &&
                    data?.map((row, index) => {
                        return (
                            <div className="collapse-child-with-count" key={index}>
                                <ConsumerWithStatus name={row.name} count={row.count} is_active={row.is_active}></ConsumerWithStatus>
                            </div>
                        );
                    })}
                {!consumerList && message ? (
                    <div className="message">
                        {header === 'Headers' && drawHeaders(data)}
                        {header === 'Payload' && (
                            <>
                                <Copy data={payload} />
                                <div className="second-row">
                                    <SegmentButton
                                        value={parser || 'string'}
                                        options={['string', 'bytes', 'json', 'protobuf']}
                                        onChange={(e) => {
                                            setParser(e);
                                            localStorage.setItem(LOCAL_STORAGE_MSG_PARSER, e);
                                        }}
                                    />
                                </div>
                                {parser === 'json' || parser === 'protobuf' ? <pre>{payload}</pre> : <p>{payload}</p>}
                            </>
                        )}
                        {header === 'Validation error' && (
                            <div className="message">
                                <p>{data}</p>
                            </div>
                        )}
                    </div>
                ) : (
                    <>
                        {!consumerList &&
                            !status &&
                            data?.length > 0 &&
                            data?.map((row, index) => {
                                return (
                                    <content is="x3d" key={index}>
                                        <p>{row.name}</p>
                                        <span>{row.value}</span>
                                    </content>
                                );
                            })}
                        {status &&
                            data?.details?.length > 0 &&
                            data?.details?.map((row, index) => {
                                return (
                                    <content is="x3d" key={index}>
                                        <p>{row.name}</p>
                                        <span>{row.value}</span>
                                    </content>
                                );
                            })}
                    </>
                )}
            </Panel>
        </Collapse>
    );
};

export default CustomCollapse;
