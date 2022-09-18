// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
