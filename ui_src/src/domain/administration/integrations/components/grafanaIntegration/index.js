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

import { INTEGRATION_LIST, getTabList } from 'const/integrationList';
import { ReactComponent as CollapseArrowIcon } from 'assets/images/collapseArrow.svg';
import grafanaps from 'assets/images/grafanaps.png';
import { ReactComponent as PurpleQuestionMark } from 'assets/images/purpleQuestionMark.svg';
import Copy from 'components/copy';
import Modal from 'components/modal';
import { ZoomInRounded } from '@material-ui/icons';
import Loader from 'components/loader';
import CustomTabs from 'components/Tabs';
import IntegrationDetails from '../integrationItem/integrationDetails';

const { Panel } = Collapse;

const ExpandIcon = ({ isActive }) => <CollapseArrowIcon className={isActive ? 'collapse-arrow open' : 'collapse-arrow close'} alt="collapse-arrow" />;

const GrafanaIntegration = ({ close }) => {
    const grafanaConfiguration = INTEGRATION_LIST['Grafana'];
    const [currentStep, setCurrentStep] = useState(0);
    const [showModal, setShowModal] = useState(false);
    const [imagesLoaded, setImagesLoaded] = useState(false);
    const [tabValue, setTabValue] = useState('Configuration');
    const tabs = getTabList('Debezium and Postgres');

    useEffect(() => {
        const images = [];
        images.push(INTEGRATION_LIST['Grafana'].banner.props.src);
        images.push(INTEGRATION_LIST['Grafana'].insideBanner.props.src);
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
    const handleToggleModal = () => {
        setShowModal(!showModal);
    };

    const getContent = (key) => {
        switch (key) {
            case 0:
                return (
                    <div className="steps-content">
                        <h3>
                            Validate that <label>Prometheus.yml</label> configfile contains "kubernetes-pods" job.
                            <br />
                            Its mandatory to scrape Memphis exporter metrics automatically.
                        </h3>
                        <div className="editor">
                            <pre>
                                {`...
honor_labels: true
        job_name: kubernetes-pods
        kubernetes_sd_configs:
        - role: pod
        relabel_configs:
        - action: keep
            regex: true
            source_labels:
            - __meta_kubernetes_pod_annotation_prometheus_io_scrape
...`}
                            </pre>
                            <Copy
                                data={`...
honor_labels: true
        job_name: kubernetes-pods
        kubernetes_sd_configs:
        - role: pod
        relabel_configs:
        - action: keep
            regex: true
            source_labels:
            - __meta_kubernetes_pod_annotation_prometheus_io_scrape
...`}
                            />
                        </div>
                    </div>
                );
            case 1:
                return (
                    <div className="steps-content">
                        <h3>
                            <b>If you haven't</b> installed Memphis with the <label>exporter.enabled</label> yet
                        </h3>
                        <div className="editor">
                            <pre>{`helm install memphis memphis 
--create-namespace --namespace memphis --wait   
--set 
cluster.enabled="true",
exporter.enabled="true"`}</pre>
                            <Copy
                                data={`helm install memphis memphis \
--create-namespace --namespace memphis --wait \
--set \
cluster.enabled="true",\
exporter.enabled="true"`}
                            />
                        </div>

                        <p>If Memphis is already installed -</p>
                        <div className="editor">
                            <pre>helm upgrade --set exporter.enabled=true memphis --namespace memphis --reuse-values</pre>
                            <Copy data={`helm upgrade --set exporter.enabled=true memphis --namespace memphis --reuse-values`} />
                        </div>
                    </div>
                );
            case 2:
                return (
                    <div className="steps-content">
                        <h3>
                            <a href="https://grafana.com/grafana/dashboards/18050-memphis-dev/" target="_blank">
                                Import
                            </a>{' '}
                            Memphis dashboard using Memphis dashboard ID: 18050
                        </h3>

                        <div className="img" onClick={handleToggleModal}>
                            <img src={grafanaps} alt="grafanaps" width={360} />
                            <ZoomInRounded style={{ color: '#ffffff', right: '45px' }} />
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
                    {grafanaConfiguration?.insideBanner}
                    <div className="integrate-header">
                        {grafanaConfiguration.header}
                        <div className="action-buttons flex-end">
                            <PurpleQuestionMark
                                className="info-icon"
                                alt="Integration info"
                                onClick={() => window.open('https://docs.memphis.dev/memphis/integrations/monitoring/grafana', '_blank')}
                            />
                        </div>
                    </div>

                    <CustomTabs value={tabValue} onChange={(tabValue) => setTabValue(tabValue)} tabs={tabs} />
                    <div className="integration-guid-body">
                        {tabValue === 'Details' && <IntegrationDetails integrateDesc={grafanaConfiguration.integrateDesc} />}
                        {tabValue === 'Configuration' && (
                            <div className="stepper-container">
                                <IntegrationDetails integrateDesc={grafanaConfiguration.integrateDesc} />
                                <div className="integration-guid-stepper">
                                    <Collapse
                                        activeKey={currentStep}
                                        onChange={(key) => setCurrentStep(Number(key))}
                                        accordion={true}
                                        expandIcon={({ isActive }) => <ExpandIcon isActive={isActive} />}
                                    >
                                        {grafanaConfiguration?.steps?.map((step) => {
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

                    {showModal && (
                        <Modal className={'zoomin-modal'} width="1000px" displayButtons={false} clickOutside={() => setShowModal(false)} open={showModal}>
                            <img width={'100%'} src={grafanaps} alt="zoomable" />
                        </Modal>
                    )}
                </>
            )}
        </dynamic-integration>
    );
};

export default GrafanaIntegration;
