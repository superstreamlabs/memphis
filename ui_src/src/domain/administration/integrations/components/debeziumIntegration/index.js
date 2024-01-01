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

import React, { useEffect, useState } from 'react';
import { useHistory } from 'react-router-dom';
import { Collapse } from 'antd';

import { INTEGRATION_LIST, getTabList } from 'const/integrationList';
import { ReactComponent as CollapseArrowIcon } from 'assets/images/collapseArrow.svg';
import { ReactComponent as PurpleQuestionMark } from 'assets/images/purpleQuestionMark.svg';
import Button from 'components/button';
import Loader from 'components/loader';
import Copy from 'components/copy';
import pathDomains from 'router';
import CustomTabs from 'components/Tabs';
import IntegrationDetails from '../integrationItem/integrationDetails';

const { Panel } = Collapse;

const ExpandIcon = ({ isActive }) => <CollapseArrowIcon className={isActive ? 'collapse-arrow open' : 'collapse-arrow close'} alt="collapse-arrow" />;

const DebeziumIntegration = ({ close }) => {
    const debeziumConfiguration = INTEGRATION_LIST['Debezium and Postgres'];
    const [currentStep, setCurrentStep] = useState(0);
    const [imagesLoaded, setImagesLoaded] = useState(false);
    const history = useHistory();
    const [tabValue, setTabValue] = useState('Configuration');
    const tabs = getTabList('Debezium and Postgres');

    useEffect(() => {
        const images = [];
        images.push(INTEGRATION_LIST['Debezium and Postgres'].banner.props.src);
        images.push(INTEGRATION_LIST['Debezium and Postgres'].insideBanner.props.src);
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

    const createNewUser = () => {
        history.push(`${pathDomains.users}`);
    };

    const getContent = (key) => {
        switch (key) {
            case 0:
                return (
                    <div className="steps-content">
                        <h3>
                            If you haven't already created an client-type Memphis user, please visit the{' '}
                            <label onClick={() => createNewUser()} style={{ cursor: 'pointer' }}>
                                User page
                            </label>{' '}
                            to create one. This user will be utilized by Debezium for specific purposes.
                        </h3>
                    </div>
                );
            case 1:
                return (
                    <div className="steps-content">
                        <h3>
                            Required Debezium configuration (normally stored in the <label>application.properties</label> file).
                        </h3>
                        <div className="editor">
                            <pre>{`debezium.sink.type=http
debezium.sink.http.url=http://<Memphis REST Gateway URL>:4444/stations/todo-cdc-events/produce/single
debezium.sink.http.time-out.ms=500
debezium.sink.http.retries=3
debezium.sink.http.authentication.type=jwt
debezium.sink.http.authentication.jwt.username=<Memphis client-type username>
debezium.sink.http.authentication.jwt.password=<Memphis client-type password>
debezium.sink.http.authentication.jwt.url=http://<Memphis REST Gateway URL>:4444/
debezium.format.key=json
debezium.format.value=json
quarkus.log.console.json=false`}</pre>
                            <Copy
                                data={`debezium.sink.type=http
                                debezium.sink.http.url=http://<Memphis REST Gateway URL>:4444/stations/todo-cdc-events/produce/single
                                debezium.sink.http.time-out.ms=500
                                debezium.sink.http.retries=3
                                debezium.sink.http.authentication.type=jwt
                                debezium.sink.http.authentication.jwt.username=<Memphis client-type username>
                                debezium.sink.http.authentication.jwt.password=<Memphis client-type password>
                                debezium.sink.http.authentication.jwt.url=http://<Memphis REST Gateway URL>:4444/
                                debezium.format.key=json
                                debezium.format.value=json
                                quarkus.log.console.json=false`}
                            />
                        </div>

                        <p>
                            In case Debezium is not installed yet, here is a quick Dockerfile to start one <br />
                            (Don't forget to enforce the config file within the container)
                        </p>
                        <div className="editor">
                            <pre>
                                {`FROM debian:bullseye-slim

RUN apt update && apt upgrade -y && apt install -y openjdk-11-jdk-headless wget git curl && rm -rf /var/cache/apt/*

WORKDIR /
RUN git clone https://github.com/debezium/debezium
WORKDIR /debezium
RUN ./mvnw clean install -DskipITs -DskipTests
WORKDIR /
RUN git clone https://github.com/debezium/debezium-server debezium-server-build
WORKDIR /debezium-server-build
RUN ./mvnw package -DskipITs -DskipTests -Passembly
RUN tar -xzvf debezium-server-dist/target/debezium-server-dist-*.tar.gz -C /
WORKDIR /debezium-server
RUN mkdir data

CMD ./run.sh`}
                            </pre>
                            <Copy
                                data={`FROM debian:bullseye-slim

RUN apt update && apt upgrade -y && apt install -y openjdk-11-jdk-headless wget git curl && rm -rf /var/cache/apt/*

WORKDIR /
RUN git clone https://github.com/debezium/debezium
WORKDIR /debezium
RUN ./mvnw clean install -DskipITs -DskipTests
WORKDIR /
RUN git clone https://github.com/debezium/debezium-server debezium-server-build
WORKDIR /debezium-server-build
RUN ./mvnw package -DskipITs -DskipTests -Passembly
RUN tar -xzvf debezium-server-dist/target/debezium-server-dist-*.tar.gz -C /
WORKDIR /debezium-server
RUN mkdir data

CMD ./run.sh`}
                            />
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
                    {debeziumConfiguration?.insideBanner}
                    <div className="integrate-header">
                        {debeziumConfiguration.header}
                        <div className="action-buttons flex-end">
                            <PurpleQuestionMark
                                className="info-icon"
                                alt="Integration info"
                                onClick={() =>
                                    window.open(
                                        'https://memphis.dev/blog/part-1-integrating-debezium-server-and-memphis-dev-for-streaming-change-data-capture-cdc-events.',
                                        '_blank'
                                    )
                                }
                            />
                        </div>
                    </div>
                    <CustomTabs value={tabValue} onChange={(tabValue) => setTabValue(tabValue)} tabs={tabs} />
                    <div className="integration-guid-body">
                        {tabValue === 'Details' && <IntegrationDetails integrateDesc={debeziumConfiguration.integrateDesc} />}
                        {tabValue === 'Configuration' && (
                            <div className="stepper-container">
                                <IntegrationDetails integrateDesc={debeziumConfiguration.integrateDesc} />
                                <div className="integration-guid-stepper">
                                    <Collapse
                                        activeKey={currentStep}
                                        onChange={(key) => setCurrentStep(Number(key))}
                                        accordion={true}
                                        expandIcon={({ isActive }) => <ExpandIcon isActive={isActive} />}
                                    >
                                        {debeziumConfiguration?.steps?.map((step) => {
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

export default DebeziumIntegration;
