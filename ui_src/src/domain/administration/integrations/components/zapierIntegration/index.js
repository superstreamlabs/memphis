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

import { INTEGRATION_LIST, getTabList } from '../../../../../const/integrationList';
import { ReactComponent as CollapseArrowIcon } from '../../../../../assets/images/collapseArrow.svg';
import { ReactComponent as PurpleQuestionMark } from '../../../../../assets/images/purpleQuestionMark.svg';
import Loader from '../../../../../components/loader';
import CustomTabs from '../../../../../components/Tabs';
import IntegrationDetails from '../integrationItem/integrationDetails';

const { Panel } = Collapse;

const ExpandIcon = ({ isActive }) => <CollapseArrowIcon className={isActive ? 'collapse-arrow open' : 'collapse-arrow close'} alt="collapse-arrow" />;

const ZapierIntegration = ({ close }) => {
    const zapierConfiguration = INTEGRATION_LIST['Zapier'];
    const [currentStep, setCurrentStep] = useState(0);
    const [imagesLoaded, setImagesLoaded] = useState(false);
    const [tabValue, setTabValue] = useState('Configuration');
    const tabs = getTabList('Zapier');

    useEffect(() => {
        const images = [];
        images.push(INTEGRATION_LIST['Zapier'].banner.props.src);
        images.push(INTEGRATION_LIST['Zapier'].insideBanner.props.src);
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
                        <h3>
                            Signing up for a free{' '}
                            <a href="https://zapier.com/apps/APP/integrations" target="_blank">
                                Zapier
                            </a>{' '}
                            account is the essential first step in using and building automation workflows.
                        </h3>
                    </div>
                );
            case 1:
                return (
                    <div className="steps-content">
                        <h3>A Zap is a workflow that connects your favorite apps to automate tasks and data transfers.</h3>
                    </div>
                );
            case 2:
                return (
                    <div className="steps-content">
                        <h3>
                            In the third step, you can harness the full potential of Zapier by integrating Memphis.dev with your workflows as either a trigger or an
                            action.
                        </h3>
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
                    {zapierConfiguration?.insideBanner}
                    <div className="integrate-header">
                        {zapierConfiguration.header}
                        <div className="action-buttons flex-end">
                            <PurpleQuestionMark
                                className="info-icon"
                                alt="Integration info"
                                onClick={() => window.open('https://docs.memphis.dev/memphis/integrations-center/processing/zapier', '_blank')}
                            />
                        </div>
                    </div>

                    <CustomTabs value={tabValue} onChange={(tabValue) => setTabValue(tabValue)} tabs={tabs} />
                    <div className="integration-guid-body">
                        {tabValue === 'Details' && <IntegrationDetails integrateDesc={zapierConfiguration.integrateDesc} />}
                        {tabValue === 'Configuration' && (
                            <div className="stepper-container">
                                <IntegrationDetails integrateDesc={zapierConfiguration.integrateDesc} />
                                <div className="integration-guid-stepper">
                                    <Collapse
                                        activeKey={currentStep}
                                        onChange={(key) => setCurrentStep(Number(key))}
                                        accordion={true}
                                        expandIcon={({ isActive }) => <ExpandIcon isActive={isActive} />}
                                    >
                                        {zapierConfiguration?.steps?.map((step) => {
                                            return (
                                                <Panel header={step.title} key={step.key}>
                                                    {getContent(step.key)}
                                                </Panel>
                                            );
                                        })}
                                    </Collapse>
                                </div>
                            </div>
                        )}
                    </div>
                </>
            )}
        </dynamic-integration>
    );
};

export default ZapierIntegration;
