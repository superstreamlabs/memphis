// Credit for The NATS.IO Authors
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
// limitations under the License.package server

import './style.scss';

import React, { useEffect, useContext, useState } from 'react';

import { LOCAL_STORAGE_ALLOW_ANALYTICS, LOCAL_STORAGE_USER_NAME } from '../../../const/localStorageConsts';
import { LOCAL_STORAGE_AVATAR_ID } from '../../../const/localStorageConsts';
import Button from '../../../components/button';
import { Context } from '../../../hooks/store';
import RadioButton from '../../../components/radioButton';
import { Checkbox } from 'antd';
import ConfImg1 from '../../../assets/images/confImg1.svg';
import ConfImg2 from '../../../assets/images/confImg2.svg';

import pathDomains from '../../../router';
import { httpRequest } from '../../../services/http';
import { ApiEndpoints } from '../../../const/apiEndpoints';
import Modal from '../../../components/modal';
import { Divider } from 'antd';

function ClusterConfiguration() {
    const [userName, setUserName] = useState('');
    const [state, dispatch] = useContext(Context);
    const [avatar, setAvatar] = useState(1);
    const [open, modalFlip] = useState(false);
    const [allowAnalytics, setAllowAnalytics] = useState();
    const [checkboxdeleteAccount, setCheckboxdeleteAccount] = useState(false);

    useEffect(() => {
        setUserName(localStorage.getItem(LOCAL_STORAGE_USER_NAME));
        setAvatar(localStorage.getItem('profile_pic') || state?.userData?.avatar_id || Number(localStorage.getItem(LOCAL_STORAGE_AVATAR_ID))); // profile_pic is available only in sandbox env
        setAllowAnalytics(localStorage.getItem(LOCAL_STORAGE_ALLOW_ANALYTICS) === 'false' ? false : true);
    }, []);

    const removeMyUser = async () => {
        try {
            await httpRequest('DELETE', `${ApiEndpoints.REMOVE_MY_UER}`);
            modalFlip(false);
            localStorage.clear();
            window.location.assign(pathDomains.login);
        } catch (err) {
            return;
        }
    };

    const sendAnalytics = async (analyticsFlag) => {
        try {
            await httpRequest('PUT', `${ApiEndpoints.EDIT_ANALYTICS}`, { send_analytics: analyticsFlag });
            setAllowAnalytics(analyticsFlag);
            localStorage.setItem(LOCAL_STORAGE_ALLOW_ANALYTICS, analyticsFlag);
        } catch (err) {
            return;
        }
    };

    return (
        <div className="configuration-container">
            <div className="header">
                <p className="main-header">Cluster configuration</p>
                <p className="sub-header">We will keep an eye on your data streams and alert you if anything went wrong according to the following triggers:</p>
            </div>
            <div className="configuration-body">
                <div className="configuration-list-container">
                    <div className="left-side">
                        <img src={ConfImg1} alt="ConfImg1" />
                        <div>
                            <p className="conf-name">ROOT_PASSWORD</p>
                            <label className="conf-description">lorem ipsumelorem ipsumelorem ipsumelorem ipsume</label>
                        </div>
                    </div>
                    <div className="right-side"></div>
                </div>
                <div className="configuration-list-container">
                    <div className="left-side">
                        <img src={ConfImg2} alt="ConfImg2" />
                        <div>
                            <p className="conf-name">POISON_MSGS_RETENTION_IN_HOURS</p>
                            <label className="conf-description">lorem ipsumelorem ipsumelorem ipsumelorem ipsume</label>
                        </div>
                    </div>
                    <div className="right-side"></div>
                </div>
                <div className="configuration-list-container">
                    <div className="left-side">
                        <img src={ConfImg1} alt="ConfImg1" />
                        <div>
                            <p className="conf-name">CONNECTION_TOKEN</p>
                            <label className="conf-description">lorem ipsumelorem ipsumelorem ipsumelorem ipsume</label>
                        </div>
                    </div>
                    <div className="right-side"></div>
                </div>
                <div className="configuration-list-container">
                    <div className="left-side">
                        <img src={ConfImg2} alt="ConfImg2" />
                        <div>
                            <p className="conf-name">POISON_MSGS_RETENTION_IN_HOURS</p>
                            <label className="conf-description">lorem ipsumelorem ipsumelorem ipsumelorem ipsume</label>
                        </div>
                    </div>
                    <div className="right-side"></div>
                </div>
            </div>
            <div className="configuration-footer">
                <div className="btn-container">
                    <Button
                        className="modal-btn"
                        width="120px"
                        height="36px"
                        placeholder="Discard"
                        colorType="black"
                        radiusType="circle"
                        backgroundColorType="none"
                        border="gray"
                        boxShadowsType="gray"
                        fontSize="14px"
                        fontWeight="600"
                        aria-haspopup="true"
                        // onClick={() => modalFlip(true)}
                    />
                    <Button
                        className="modal-btn"
                        width="180px"
                        height="36px"
                        placeholder="Save changes"
                        colorType="white"
                        radiusType="circle"
                        backgroundColorType="purple"
                        border="red"
                        boxShadowsType="red"
                        fontSize="14px"
                        fontWeight="600"
                        aria-haspopup="true"
                        // onClick={() => modalFlip(true)}
                    />
                </div>
            </div>
        </div>
    );
}

export default ClusterConfiguration;
