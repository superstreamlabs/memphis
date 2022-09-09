// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
import './style.scss';

import React, { useState } from 'react';
import { Collapse } from 'antd';

import CollapseArrow from '../../../../../assets/images/collapseArrow.svg';
import StatusIndication from '../../../../../components/indication';

const { Panel } = Collapse;

const MultiCollapse = ({ data, header, defaultOpen }) => {
    const [activeKey, setActiveKey] = useState(defaultOpen ? ['1'] : []);
    const [activeChiledKey, setActiveChiledKey] = useState();

    const onChange = (key) => {
        setActiveKey(key);
    };
    const onChiledChange = (key) => {
        setActiveChiledKey(key);
    };

    return (
        <>
            {header !== undefined ? (
                <Collapse ghost defaultActiveKey={activeKey} onChange={onChange} className="custom-collapse multi">
                    <Panel
                        showArrow={false}
                        collapsible={data?.length === 0 ? 'disabled' : null}
                        header={
                            <div className="collapse-header">
                                <p className="title">
                                    {header} <span className="consumer-number">{data?.length}</span>
                                </p>

                                <status is="x3d">
                                    <img className={activeKey[0] === '1' ? 'collapse-arrow open' : 'collapse-arrow close'} src={CollapseArrow} alt="collapse-arrow" />
                                </status>
                            </div>
                        }
                        key="1"
                    >
                        <Collapse ghost accordion={true} className="collapse-child" onChange={onChiledChange}>
                            {data?.length > 0 &&
                                data?.map((row, index) => {
                                    return (
                                        <Panel
                                            showArrow={false}
                                            header={
                                                <div className="collapse-header">
                                                    <p className="title">{row.name}</p>
                                                    <status is="x3d">
                                                        <StatusIndication is_active={row.is_active} is_deleted={row.is_deleted} />
                                                        <img
                                                            className={Number(activeChiledKey) === index ? 'collapse-arrow open' : 'collapse-arrow close'}
                                                            src={CollapseArrow}
                                                            alt="collapse-arrow"
                                                        />
                                                    </status>
                                                </div>
                                            }
                                            key={index}
                                        >
                                            {row.details?.length > 0 &&
                                                row.details?.map((row, index) => {
                                                    return (
                                                        <div className="panel-child" key={index}>
                                                            <content is="x3d" key={index}>
                                                                <p>{row.name}</p>
                                                                <span>{row.value}</span>
                                                            </content>
                                                        </div>
                                                    );
                                                })}
                                        </Panel>
                                    );
                                })}
                        </Collapse>
                    </Panel>
                </Collapse>
            ) : (
                <div className="custom-collapse multi">
                    <Collapse ghost accordion={true} className="collapse-child" onChange={onChiledChange}>
                        {data?.length > 0 &&
                            data?.map((row, index) => {
                                return (
                                    <Panel
                                        showArrow={false}
                                        header={
                                            <div className="collapse-header">
                                                <p className="title">{row.name}</p>
                                                <status is="x3d">
                                                    <StatusIndication is_active={row.is_active} is_deleted={row.is_deleted} />
                                                    <img
                                                        className={Number(activeChiledKey) === index ? 'collapse-arrow open' : 'collapse-arrow close'}
                                                        src={CollapseArrow}
                                                        alt="collapse-arrow"
                                                    />
                                                </status>
                                            </div>
                                        }
                                        key={index}
                                    >
                                        {row.details?.length > 0 &&
                                            row.details?.map((row, index) => {
                                                return (
                                                    <div className="panel-child" key={index}>
                                                        <content is="x3d" key={index}>
                                                            <p>{row.name}</p>
                                                            <span>{row.value}</span>
                                                        </content>
                                                    </div>
                                                );
                                            })}
                                    </Panel>
                                );
                            })}
                    </Collapse>
                </div>
            )}
        </>
    );
};

export default MultiCollapse;
