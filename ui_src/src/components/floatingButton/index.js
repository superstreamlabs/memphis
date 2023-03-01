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

import React, { useState } from 'react';
import Modal from '../modal';
import Installation from '../installation';
import installationIcon from '../../assets/images/installationIcon.svg';
import Draggable from 'react-draggable';
import CloudDownloadIcon from '@material-ui/icons/CloudDownload';
import ExpandLessIcon from '@material-ui/icons/ExpandLess';

const FloatingButton = () => {
    const [showInstallaion, setShowInstallaion] = useState(false);
    const [expendBox, setexpendBoxBox] = useState(true);

    const openModal = () => {
        setexpendBoxBox(false);
        setShowInstallaion(true);
    };

    return (
        <div className="floating-button-container">
            <Draggable defaultPosition={{ x: 0, y: 600 }} bounds="body" axis="y">
                <div>
                    <div className={!expendBox ? 'box-wrapper' : 'box-wrapper open'} onClick={() => (!expendBox ? setexpendBoxBox(true) : null)}>
                        {expendBox && (
                            <>
                                <div className="close-box" onClick={() => setexpendBoxBox(false)}>
                                    <ExpandLessIcon />
                                </div>
                                <div className="box-open">
                                    <CloudDownloadIcon />
                                    <p onClick={openModal}>Install Now</p>
                                    <span onClick={openModal}>Choose an enviroment ></span>
                                </div>
                            </>
                        )}
                        {!expendBox && <CloudDownloadIcon className="download-icon" />}
                    </div>
                </div>
            </Draggable>
            <Modal
                header={
                    <label className="installation-icon-wrapper">
                        <img src={installationIcon} alt="installationIcon" />
                    </label>
                }
                height="700px"
                clickOutside={() => {
                    setShowInstallaion(false);
                }}
                open={showInstallaion}
                displayButtons={false}
            >
                <Installation closeModal={() => setShowInstallaion(false)} />
            </Modal>
        </div>
    );
};

export default FloatingButton;
