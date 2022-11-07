// Credit for The NATS.IO Authors
// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package server
import './style.scss';

import React, { useState } from 'react';
import { Collapse } from 'antd';

import CollapseArrow from '../../../../../assets/images/collapseArrow.svg';
import StatusIndication from '../../../../../components/indication';
import Copy from '../../../../../components/copy';
import OverflowTip from '../../../../../components/tooltip/overflowtip';

const { Panel } = Collapse;

const CustomCollapse = ({ status, data, header, defaultOpen, message }) => {
    const [activeKey, setActiveKey] = useState(defaultOpen ? ['1'] : []);
    const onChange = (key) => {
        setActiveKey(key);
    };

    const isJsonString = (str) => {
        try {
            JSON.parse(str);
        } catch (e) {
            return false;
        }
        return true;
    };

    const drawHeaders = (headers) => {
        let obj = [];
        for (const property in headers) {
            obj.push(
                <div className="headers-container">
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
                header={
                    <div className="collapse-header">
                        <p className="title">{header}</p>
                        <status is="x3d">
                            {/* {status && <StatusIndication is_active={data?.is_active} is_deleted={data?.is_deleted} />} */}
                            <img className={activeKey[0] === '1' ? 'collapse-arrow open' : 'collapse-arrow close'} src={CollapseArrow} alt="collapse-arrow" />
                        </status>
                    </div>
                }
                key="1"
            >
                {message ? (
                    <div className="message">
                        {header === 'Payload' && isJsonString(data) && <pre>{JSON.stringify(JSON.parse(data), null, 2)}</pre>}
                        {header === 'Payload' && !isJsonString(data) && <p>{data}</p>}
                        {header === 'Headers' && drawHeaders(data)}
                    </div>
                ) : (
                    <>
                        {!status &&
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
