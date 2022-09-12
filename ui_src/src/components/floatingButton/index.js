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
// limitations under the License.

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
                        <img src={installationIcon} />
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
