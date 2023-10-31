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

import React from 'react';
import { ReactComponent as FunctionIntegrateIcon } from '../../assets/images/functionIntegrate.svg';
import BundleBanner from '../../assets/images/banners/bundle1.png';
import CloudBanner from '../../assets/images/banners/cloud2.png';
import FunctionsBanner from '../../assets/images/banners/function3.png';
import UpgradeBanner from '../../assets/images/banners/upgrade4.png';
import UpgradePlans from '../upgradePlans';
import Modal from '../modal';
import Button from '../button';

const CloudModal = ({ type, open, handleClose }) => {
    const content = {
        bundle: {
            title: <label className="cloud-gradient">Enhance Your Journey</label>,
            subtitle: 'Get Your Open-Source Support Bundle Today!',
            banner: BundleBanner,
            leftBtn: 'Learn More',
            leftBtnLink: 'https://docs.memphis.dev/memphis/open-source-installation/open-source-support-bundle/',
            rightBtn: 'Book a Call',
            rightBtnLink: 'https://meetings.hubspot.com/yaniv-benhemo'
        },
        cloud: {
            title: <label className="cloud-gradient">Enhance Your Journey</label>,
            subtitle: 'Embrace serverless, enjoy peace of mind, and experience enhanced resilience.',
            banner: CloudBanner,
            leftBtn: 'Learn More',
            leftBtnLink: 'https://memphis.dev/memphis-dev-cloud/',
            rightBtn: 'Claim 50% Discount',
            rightBtnLink: 'https://meetings.hubspot.com/yaniv-benhemo'
        },
        upgrade: {
            title: (
                <>
                    <label className="cloud-gradient">Upgrade</label>
                    <label className="title-label"> your plan</label>
                </>
            ),
            subtitle: 'To Unlock More Features And Enhance Your Experience!',
            banner: UpgradeBanner,
            leftBtn: 'Upgrade Now',
            leftBtnLink: '',
            rightBtn: 'Talk to Sales',
            rightBtnLink: 'https://meetings.hubspot.com/yaniv-benhemo'
        },
        functions: {
            title: (
                <>
                    <label className="title-label">The Future of </label>
                    <label className="cloud-gradient">Event-Driven</label>
                </>
            ),
            subtitle: 'Discover A Faster And Smarter Way To Do Event-driven And Stream Processing',
            banner: FunctionsBanner,
            leftBtn: 'Learn More',
            leftBtnLink: 'https://memphis.dev/memphis-dev-cloud/',
            rightBtn: 'Book a demo',
            rightBtnLink: 'https://meetings.hubspot.com/yaniv-benhemo/demo-for-memphis-functions'
        }
    };

    return (
        <cloud-modal is="x3d">
            <Modal
                header={
                    <div
                        style={{
                            display: 'flex',
                            flexDirection: 'column',
                            justifyContent: 'center',
                            alignItems: 'center',
                            fontFamily: 'InterSemiBold',
                            fontSize: '16px',
                            margin: 0
                        }}
                    >
                        <div className="header-icon">
                            <FunctionIntegrateIcon width={22} height={22} />
                        </div>
                        <span>{content[type]?.title}</span>
                        <label style={{ display: 'flex', justifyContent: 'center', textAlign: 'center', fontFamily: 'Inter', fontSize: '14px' }}>
                            {content[type]?.subtitle}
                        </label>
                    </div>
                }
                displayButtons={false}
                width="685px"
                clickOutside={handleClose}
                open={open}
                className="cloud-modal"
            >
                <div className="banner-container">
                    <img src={content[type]?.banner} alt="banner" className="banner" />
                </div>
                <span className="cloud-modal-btns">
                    {type === 'upgrade' ? (
                        <UpgradePlans
                            content={
                                <Button
                                    width="230px"
                                    height="40px"
                                    placeholder={content[type]?.leftBtn}
                                    colorType="black"
                                    radiusType="circle"
                                    backgroundColorType={'white'}
                                    border={'gray'}
                                    fontSize="12px"
                                    fontWeight="bold"
                                    onClick={() => type !== 'upgrade' && window.open(content[type]?.leftBtnLink, '_blank')}
                                />
                            }
                            isExternal={false}
                        />
                    ) : (
                        <Button
                            width="230px"
                            height="40px"
                            placeholder={content[type]?.leftBtn}
                            colorType="black"
                            radiusType="circle"
                            backgroundColorType={'white'}
                            border={'gray'}
                            fontSize="12px"
                            fontWeight="bold"
                            onClick={() => window.open(content[type]?.leftBtnLink, '_blank')}
                        />
                    )}
                    <Button
                        width="230px"
                        height="40px"
                        placeholder={content[type]?.rightBtn}
                        colorType="white"
                        radiusType="circle"
                        backgroundColorType={'purple'}
                        border={'gray'}
                        fontSize="12px"
                        fontWeight="bold"
                        onClick={() => window.open(content[type]?.rightBtnLink, '_blank')}
                    />
                </span>
            </Modal>
        </cloud-modal>
    );
};

export default CloudModal;
