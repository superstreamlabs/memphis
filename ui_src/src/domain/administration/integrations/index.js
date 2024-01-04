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

import React, { useEffect, useContext, useState } from 'react';
import { ReactComponent as IntegrationRequestIcon } from 'assets/images/integrationRequestIcon.svg';
import { CATEGORY_LIST, INTEGRATION_LIST } from 'const/integrationList';
import { ReactComponent as SoonBadgeIcon } from 'assets/images/soonBadge.svg';
import IntegrationItem from './components/integrationItem';
import { ApiEndpoints } from 'const/apiEndpoints';
import { isCloud } from 'services/valueConvertor';
import { httpRequest } from 'services/http';
import Button from 'components/button';
import Loader from 'components/loader';
import { Context } from 'hooks/store';
import Modal from 'components/modal';
import Input from 'components/Input';
import Tag from 'components/tag';
import { showMessages } from 'services/genericServices';
import { useLocation } from 'react-router-dom';
import {entitlementChecker} from "utils/plan";

const Integrations = () => {
    const [state, dispatch] = useContext(Context);
    const [modalIsOpen, modalFlip] = useState(false);
    const [integrationRequest, setIntegrationRequest] = useState('');
    const [categoryFilter, setCategoryFilter] = useState('All');
    const [filterList, setFilterList] = useState(INTEGRATION_LIST);
    const [imagesLoaded, setImagesLoaded] = useState(false);
    const [githubModalIsOpen, setGithubModalIsOpen] = useState(false);
    const location = useLocation();
    const queryParameters = new URLSearchParams(location.search);

    const storageTiringLimits = !entitlementChecker(state, 'feature-storage-tiering');

    useEffect(() => {
        const process = async () => {
            const installationId = queryParameters.get('installation_id');
            if (installationId) {
                window.history.replaceState({}, null, location.pathname);
                try {
                    const res = await httpRequest('POST', ApiEndpoints.CREATE_INTEGRATION, { name: 'github', keys: { installation_id: installationId } });
                    setTimeout(async () => {
                        dispatch({ type: 'ADD_INTEGRATION', payload: res });
                        await getallIntegration();
                        setGithubModalIsOpen(true);
                    }, 2000);
                } catch (err) {
                    console.log(err);
                }
            } else {
                getallIntegration();
            }
        };
        process();
    }, []);

    useEffect(() => {
        const images = [];
        Object.values(INTEGRATION_LIST).forEach((integration) => {
            images.push(integration.banner.props.src);
            images.push(integration.insideBanner.props.src);
        });
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

    useEffect(() => {
        switch (categoryFilter) {
            case 'All':
                setFilterList(INTEGRATION_LIST);
                break;
            default:
                let filteredList = Object.values(INTEGRATION_LIST).filter((integration) => integration.category.name === categoryFilter);
                setFilterList(filteredList);
                break;
        }
    }, [categoryFilter]);

    const getallIntegration = async () => {
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_ALL_INTEGRATION);
            dispatch({ type: 'SET_INTEGRATIONS', payload: data || [] });
        } catch (err) {
            return;
        }
    };
    const handleSendRequest = async () => {
        try {
            await httpRequest('POST', ApiEndpoints.REQUEST_INTEGRATION, { request_content: integrationRequest });
            showMessages('success', 'Thanks for your feedback');
            modalFlip(false);
            setIntegrationRequest('');
        } catch (err) {
            return;
        }
    };

    return (
        <div className="alerts-integrations-container">
            <div className="header-preferences">
                <div className="left">
                    <p className="main-header">Integrations center</p>
                    <p className="memphis-label">Integrations for notifications, monitoring, API calls, and more</p>
                </div>
                <Button
                    width="180px"
                    height="35px"
                    placeholder="Request a new integration"
                    colorType="white"
                    radiusType="circle"
                    backgroundColorType="purple"
                    border="none"
                    fontSize="12px"
                    fontFamily="InterSemiBold"
                    onClick={() => modalFlip(true)}
                />
            </div>
            <div className="categories-list">
                {Object.keys(CATEGORY_LIST).map((key) => {
                    const category = CATEGORY_LIST[key];
                    const isCategoryFilter = categoryFilter === category.name;
                    return <Tag key={key} tag={category} onClick={(e) => setCategoryFilter(e)} border={isCategoryFilter} />;
                })}
            </div>
            {!imagesLoaded && (
                <div className="loading">
                    <Loader background={false} />
                </div>
            )}
            {imagesLoaded && (
                <div className="integration-list">
                    {Object.keys(filterList)?.map((integration) => {
                        const integrationItem = filterList[integration];
                        const key = integrationItem.name;
                        const integrationElement = (
                            <IntegrationItem
                                lockFeature={isCloud() && integrationItem.name === 'S3' && storageTiringLimits}
                                key={key}
                                value={integrationItem}
                                isOpen={integration === 'GitHub' ? githubModalIsOpen : false}
                            />
                        );

                        if (integrationItem.comingSoon) {
                            return (
                                <div key={key} className="cloud-wrapper">
                                    <div className="dark-background">
                                        <SoonBadgeIcon className="cloud-badge" alt="cloud badge" />
                                    </div>
                                    {integrationElement}
                                </div>
                            );
                        }
                        return integrationElement;
                    })}
                </div>
            )}
            <Modal
                className="request-integration-modal"
                header={<IntegrationRequestIcon alt="errorModal" />}
                height="250px"
                width="450px"
                displayButtons={false}
                clickOutside={() => modalFlip(false)}
                open={modalIsOpen}
            >
                <div className="roll-back-modal">
                    <p className="title">Integrations framework</p>
                    <p className="desc">Until our integrations framework will be released, we can build it for you. Which integration is missing?</p>
                    <Input
                        placeholder="App & reason"
                        type="text"
                        fontSize="12px"
                        radiusType="semi-round"
                        colorType="black"
                        backgroundColorType="none"
                        borderColorType="gray"
                        height="40px"
                        onBlur={(e) => setIntegrationRequest(e.target.value)}
                        onChange={(e) => setIntegrationRequest(e.target.value)}
                        value={integrationRequest}
                    />

                    <div className="buttons">
                        <Button
                            width="150px"
                            height="34px"
                            placeholder="Cancel"
                            colorType="black"
                            radiusType="circle"
                            backgroundColorType="white"
                            border="gray-light"
                            fontSize="12px"
                            fontFamily="InterSemiBold"
                            onClick={() => modalFlip(false)}
                        />
                        <Button
                            width="150px"
                            height="34px"
                            placeholder="Send"
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="12px"
                            fontFamily="InterSemiBold"
                            disabled={integrationRequest === ''}
                            onClick={() => handleSendRequest()}
                        />
                    </div>
                </div>
            </Modal>
        </div>
    );
};

export default Integrations;
