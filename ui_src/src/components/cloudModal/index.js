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
    return (
        <cloud-modal is="x3d">
            <Modal
                header={
                    <div>
                        {type === 'oss' ? (
                            <>
                                <span style={{ display: 'flex', justifyContent: 'center', fontFamily: 'InterSemiBold', fontSize: '16px', margin: 0 }}>
                                    <label>Elevate Your Experience with Memphis.dev </label>
                                    <label className="cloud-gradient">Cloud</label>
                                    <label>!</label>
                                </span>
                                <label style={{ display: 'flex', justifyContent: 'center', textAlign: 'center', fontFamily: 'Inter', fontSize: '14px' }}>
                                    Embrace serverless, enjoy peace of mind, and experience enhanced resilience.
                                </label>
                            </>
                        ) : (
                            <span>cloud</span>
                        )}
                    </div>
                }
                displayButtons={false}
                width="550px"
                height="380px"
                clickOutside={handleClose}
                open={open}
                className="cloud-modal"
            >
                <img src={CloudBanner} className="banner" alt="benner" />
                <span className="cloud-modal-btns">
                    <Button
                        width="230px"
                        height="40px"
                        placeholder="Learn More"
                        colorType="black"
                        radiusType="circle"
                        backgroundColorType={'white'}
                        border={'gray'}
                        fontSize="12px"
                        fontWeight="bold"
                        onClick={() => window.open('https://memphis.dev/memphis-dev-cloud/', '_blank')}
                    />
                    <Button
                        width="230px"
                        height="40px"
                        placeholder="Schedule a call"
                        colorType="white"
                        radiusType="circle"
                        backgroundColorType={'purple'}
                        border={'gray'}
                        fontSize="12px"
                        fontWeight="bold"
                        onClick={() =>
                            type === 'oss'
                                ? window.open('https://meetings.hubspot.com/yaniv-benhemo', '_blank')
                                : window.open('https://memphisdev.github.io/memphis/', '_blank')
                        }
                    />
                </span>
            </Modal>
        </cloud-modal>
    );
};

export default CloudModal;
