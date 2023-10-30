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
import CloudBanner from '../../assets/images/banners/cloudBanner.jpg';
import Modal from '../modal';
import Button from '../button';

const CloudModal = ({ type, open, handleClose }) => {
    const content = {
        oss: {
            title: (
                <>
                    <label>Elevate Your Experience with Memphis.dev </label>
                    <label className="cloud-gradient">Cloud</label>
                    <label>!</label>
                </>
            ),
            subtitle: 'Embrace serverless, enjoy peace of mind, and experience enhanced resilience.',
            banner: CloudBanner,
            leftBtn: 'Learn More',
            leftBtnLink: 'https://memphis.dev/memphis-dev-cloud/',
            rightBtn: 'Claim a 50% discount',
            rightBtnLink: 'https://meetings.hubspot.com/yaniv-benhemo'
        },
        cloud: {
            title: 'Elevate Your Experience with Memphis.dev Cloud!',
            subtitle: 'Embrace serverless, enjoy peace of mind, and experience enhanced resilience.',
            banner: CloudBanner,
            leftBtn: 'Learn More',
            leftBtnLink: 'https://memphis.dev/memphis-dev-cloud/',
            rightBtn: 'Schedule a call',
            rightBtnLink: 'https://meetings.hubspot.com/yaniv-benhemo'
        },
        upgrade: {
            title: 'Elevate Your Experience with Memphis.dev Cloud!',
            subtitle: 'Embrace serverless, enjoy peace of mind, and experience enhanced resilience.',
            banner: CloudBanner,
            leftBtn: 'Learn More',
            leftBtnLink: 'https://memphis.dev/memphis-dev-cloud/',
            rightBtn: 'Schedule a call',
            rightBtnLink: 'https://meetings.hubspot.com/yaniv-benhemo'
        }
    };

    return (
        <cloud-modal is="x3d">
            <Modal
                header={
                    <div>
                        <span style={{ display: 'flex', justifyContent: 'center', fontFamily: 'InterSemiBold', fontSize: '16px', margin: 0 }}>
                            {content[type]?.title}
                        </span>
                        <label style={{ display: 'flex', justifyContent: 'center', textAlign: 'center', fontFamily: 'Inter', fontSize: '14px' }}>
                            {content[type]?.subtitle}
                        </label>
                    </div>
                }
                displayButtons={false}
                width="550px"
                height="380px"
                clickOutside={handleClose}
                open={open}
                className="cloud-modal"
            >
                <img src={content[type]?.banner} className="banner" alt="benner" />
                <span className="cloud-modal-btns">
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
