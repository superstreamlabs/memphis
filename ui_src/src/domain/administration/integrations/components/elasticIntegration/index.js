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
import { Collapse } from 'antd';

import { INTEGRATION_LIST } from '../../../../../const/integrationList';
import { ReactComponent as CollapseArrowIcon } from '../../../../../assets/images/collapseArrow.svg';

import Button from '../../../../../components/button';
import Copy from '../../../../../components/copy';
import Loader from '../../../../../components/loader';

const { Panel } = Collapse;

const ExpandIcon = ({ isActive }) => <CollapseArrowIcon className={isActive ? 'collapse-arrow open' : 'collapse-arrow close'} alt="collapse-arrow" />;

const ElasticIntegration = ({ close }) => {
    const elasticConfiguration = INTEGRATION_LIST['Elasticsearch'];
    const [currentStep, setCurrentStep] = useState(0);
    const [imagesLoaded, setImagesLoaded] = useState(false);

    useEffect(() => {
        const images = [];
        images.push(INTEGRATION_LIST['Elasticsearch'].banner.props.src);
        images.push(INTEGRATION_LIST['Elasticsearch'].insideBanner.props.src);
        images.push(INTEGRATION_LIST['Elasticsearch'].icon.props.src);
        const promises = [];

        images.forEach((imageUrl) => {
            const image = new Image();
            promises.push(
                new Promise((resolve) => {
                    image.onload = resolve;
                })
            );
            image.src = imageUrl;
        });

        Promise.all(promises).then(() => {
            setImagesLoaded(true);
        });
    }, []);

    const getContent = (key) => {
        switch (key) {
            case 0:
                return (
                    <div className="steps-content">
                        <h3>Download the manifest file:</h3>
                        <div className="editor">
                            <pre>curl -L -O https://raw.githubusercontent.com/elastic/elastic-agent/master/deploy/kubernetes/elastic-agent-managed-kubernetes.yaml</pre>
                            <Copy data="curl -L -O https://raw.githubusercontent.com/elastic/elastic-agent/master/deploy/kubernetes/elastic-agent-managed-kubernetes.yaml" />
                        </div>
                    </div>
                );
            case 1:
                return (
                    <div className="steps-content">
                        <h3>
                            The Elastic Agent needs to be assigned to a policy to enable the proper inputs. To achieve Kubernetes observability, the policy needs to
                            include the Kubernetes integration. Refer to{' '}
                            <a href="https://www.elastic.co/guide/en/fleet/master/agent-policy.html#create-a-policy" target="_blank">
                                Create a policy
                            </a>{' '}
                            and{' '}
                            <a href="https://www.elastic.co/guide/en/fleet/master/agent-policy.html#add-integration" target="_blank">
                                Add an integration to a policy
                            </a>{' '}
                            to learn how to configure the{' '}
                            <a href="https://docs.elastic.co/en/integrations/kubernetes" target="_blank">
                                Kubernetes integration
                            </a>
                            .
                        </h3>
                    </div>
                );
            case 2:
                return (
                    <div className="steps-content">
                        <h3>Enrollment of an Elastic Agent is defined as the action to register a specific agent to a running Fleet Server within the manifest file.</h3>
                    </div>
                );
            case 3:
                return (
                    <div className="steps-content">
                        <div className="editor">
                            <pre>kubectl create -f elastic-agent-managed-kubernetes.yaml</pre>
                            <Copy data="kubectl create -f elastic-agent-managed-kubernetes.yaml" />
                        </div>
                    </div>
                );
            default:
                break;
        }
    };

    return (
        <dynamic-integration is="3xd" className="integration-modal-container">
            {!imagesLoaded && (
                <div className="loader-integration-box">
                    <Loader />
                </div>
            )}
            {imagesLoaded && (
                <>
                    {elasticConfiguration?.insideBanner}
                    <div className="integrate-header">
                        {elasticConfiguration.header}
                        <div className="action-buttons flex-end">
                            <Button
                                width="140px"
                                height="35px"
                                placeholder="Integration guide"
                                colorType="white"
                                radiusType="circle"
                                backgroundColorType="purple"
                                border="none"
                                fontSize="12px"
                                fontFamily="InterSemiBold"
                                onClick={() => window.open('https://docs.memphis.dev/memphis/integrations/monitoring/elasticsearch-observability', '_blank')}
                            />
                        </div>
                    </div>
                    {elasticConfiguration.integrateDesc}
                    <div className="integration-guid-stepper">
                        <Collapse
                            activeKey={currentStep}
                            onChange={(key) => setCurrentStep(Number(key))}
                            accordion={true}
                            expandIcon={({ isActive }) => <ExpandIcon isActive={isActive} />}
                        >
                            {elasticConfiguration?.steps?.map((step) => {
                                return (
                                    <Panel header={step.title} key={step.key}>
                                        {getContent(step.key)}
                                    </Panel>
                                );
                            })}
                        </Collapse>
                        <div className="close-btn">
                            <Button
                                width="300px"
                                height="45px"
                                placeholder="Close"
                                colorType="white"
                                radiusType="circle"
                                backgroundColorType="purple"
                                fontSize="14px"
                                fontFamily="InterSemiBold"
                                onClick={() => close()}
                            />
                        </div>
                    </div>
                </>
            )}
        </dynamic-integration>
    );
};

export default ElasticIntegration;
