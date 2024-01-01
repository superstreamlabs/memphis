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

import React, { useContext, useState, useEffect } from 'react';
import SysContainers from './sysContainers';
import Component from './components/component';
import { Context } from 'hooks/store';
import { Tree } from 'antd';
import { ReactComponent as CollapseArrowIcon } from 'assets/images/collapseArrow.svg';

const SysComponents = () => {
    const [state, dispatch] = useContext(Context);
    const [expandedNodes, setExpandedNodes] = useState(['0-0']);

    const getBrokers = (comp) => {
        const children = [];

        ['unhealthy_components', 'dangerous_components', 'risky_components', 'healthy_components'].forEach((type) => {
            const typeComponents = comp.components[type];
            if (typeComponents) {
                children.push(
                    ...typeComponents
                        .sort((a, b) => {
                            return a.name.split('-')[1] - b.name.split('-')[1];
                        })
                        .map((component, index) => ({
                            title: (
                                <SysContainers
                                    key={`0-${type}-${index}`}
                                    component={component}
                                    k8sEnv={state?.monitor_data?.k8s_env}
                                    metricsEnabled={state?.monitor_data?.metrics_enabled}
                                    index={index}
                                />
                            ),
                            key: `0-${type}-${index}`,
                            selectable: false
                        }))
                );
            }
        });
        return children;
    };

    return (
        <div className="overview-components-wrapper system-components-wrapper">
            <div className="system-components-container">
                <div className="overview-components-header">
                    <p>System components</p>
                    <label>An essential components list of your system</label>
                </div>
                <div className="component-list">
                    {state?.monitor_data?.system_components?.map((comp, i) => {
                        const childrenData = getBrokers(comp);
                        return (
                            <Tree
                                key={`tree-node${i}`}
                                blockNode
                                // showLine={childrenData?.length > 0} //currently disabled due to bug in antd
                                selectable={childrenData?.length > 0}
                                expandedKeys={expandedNodes}
                                switcherIcon={({ expanded }) =>
                                    childrenData.length > 0 && <CollapseArrowIcon className={expanded ? 'collapse-arrow open' : 'collapse-arrow'} alt="collapse-arrow" />
                                }
                                rootClassName={!expandedNodes?.includes(`0-${i}`) && 'divided'}
                                onSelect={(_, info) => {
                                    if (!expandedNodes?.includes(info.node.key)) setExpandedNodes([...expandedNodes, info.node.key]);
                                    else setExpandedNodes(expandedNodes.filter((node) => node !== info.node.key));
                                }}
                                defaultExpandedKeys={childrenData?.length > 0 ? ['0-0'] : []}
                                onExpand={(_, { expanded }) => {
                                    if (expanded) setExpandedNodes([...expandedNodes, `0-${i}`]);
                                    else setExpandedNodes(expandedNodes.filter((node) => node !== `0-${i}`));
                                }}
                                treeData={[
                                    {
                                        title: <Component comp={comp} i={i} />,
                                        key: `0-${i}`,
                                        children: childrenData
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
