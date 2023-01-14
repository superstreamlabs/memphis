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

import React, { useContext, useEffect } from 'react';
import SysContainers from './sysContainers';
import Component from './component';
import { Context } from '../../../hooks/store';
import { Tree } from 'antd';
import CollapseArrow from '../../../assets/images/collapseArrow.svg';

const SysComponents = () => {
    const [state, dispatch] = useContext(Context);

    return (
        <div className="overview-wrapper">
            <div className="system-components-container">
                <div className="overview-components-header">
                    <p>System components</p>
                </div>
                <div className="component-list">
                    {state?.monitor_data?.system_components?.map((comp, i) => {
                        return (
                            <Tree
                                blockNode
                                showLine
                                switcherIcon={({ expanded }) => (
                                    <img className={expanded ? 'collapse-arrow open' : 'collapse-arrow'} src={CollapseArrow} alt="collapse-arrow" />
                                )}
                                defaultExpandedKeys={['0-0-0']}
                                treeData={[
                                    {
                                        title: <Component comp={comp} i={i} />,
                                        key: `0-${i}`,
                                        children: comp?.components?.map((component) => {
                                            return {
                                                title: <SysContainers component={component} />,
                                                key: `0-${i}-0`,
                                                selectable: false
                                            };
                                        })
                                    }
                                ]}
                            />
                        );
                    })}
                </div>
            </div>
        </div>
    );
};

export default SysComponents;
