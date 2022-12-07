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
import { Context } from '../../../hooks/store';
import pathDomains from '../../../router';
import { LOCAL_STORAGE_ALLOW_ANALYTICS, LOCAL_STORAGE_USER_NAME, LOCAL_STORAGE_COMPANY_LOGO, LOCAL_STORAGE_AVATAR_ID } from '../../../const/localStorageConsts';
import { Checkbox, Divider, Upload, message } from 'antd';
import RadioButton from '../../../components/radioButton';
import Button from '../../../components/button';
import Modal from '../../../components/modal';
import { ApiEndpoints } from '../../../const/apiEndpoints';
import { httpRequest } from '../../../services/http';
import ImgUploader from './imgUploader';

function Profile() {
    const [userName, setUserName] = useState('');
    const [state, dispatch] = useContext(Context);
    const [avatar, setAvatar] = useState(1);
    const [open, modalFlip] = useState(false);
    const [allowAnalytics, setAllowAnalytics] = useState();
    const [checkboxdeleteAccount, setCheckboxdeleteAccount] = useState(false);

    useEffect(() => {
        setUserName(localStorage.getItem(LOCAL_STORAGE_USER_NAME));
        setAvatar(Number(localStorage.getItem(LOCAL_STORAGE_AVATAR_ID)) || state?.userData?.avatar_id);
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

    const editAvatar = async (avatarId) => {
        try {
            const data = await httpRequest('PUT', `${ApiEndpoints.EDIT_AVATAR}`, { avatar_id: avatarId });
            setAvatar(data.avatar_id);
            localStorage.setItem(LOCAL_STORAGE_AVATAR_ID, data.avatar_id);
            dispatch({ type: 'SET_AVATAR_ID', payload: data.avatar_id });
        } catch (err) {
            return;
        }
    };

    return (
        <div className="profile-container">
            <div className="header-preferences">
                <p className="main-header">Profile</p>
                <p className="sub-header">Modify your profile information and preferences</p>
            </div>
            <div className="avatar-section">
                <p className="title">Avatar</p>
                <div className="avatar-images">
                    {process.env.REACT_APP_SANDBOX_ENV && localStorage.getItem('profile_pic') && (
                        <div className={'avatar-img selected'}>
                            <img src={localStorage.getItem('profile_pic')} width={35} height={35} alt="avater" />
                        </div>
                    )}
                    {Array.from(Array(8).keys()).map((item, index) => {
                        return (
                            <div
                                key={index}
                                className={
                                    process.env.REACT_APP_SANDBOX_ENV && localStorage.getItem('profile_pic')
                                        ? 'avatar-img avatar-disable'
                                        : avatar === item + 1
                                        ? 'avatar-img selected'
                                        : 'avatar-img'
                                }
                                onClick={process.env.REACT_APP_SANDBOX_ENV ? '' : () => editAvatar(item + 1)}
                            >
                                <img src={require(`../../../assets/images/bots/avatar${item + 1}.svg`)} alt="avater" />
                            </div>
                        );
                    })}
                </div>
            </div>
            <ImgUploader />
            <Divider />
            <div className="analytics-section">
                <p className="title">Analytics</p>
                <label className="analytics-description">
                    Memphis only collects bugs, events, and anonymous metadata to become better and more stable for you.
                    <br />
                    No sensitive or personal data gets collected.
                </label>
                <div className="radioButton-section">
                    <RadioButton
                        options={[
                            { id: 0, value: true, label: 'Allow Analytics' },
                            { id: 1, value: false, label: 'Don’t allow any analytics' }
                        ]}
                        radioValue={allowAnalytics}
                        onChange={(e) => sendAnalytics(e.target.value)}
                        onClick={(e) => sendAnalytics(e)}
                        labelType
                    />
                </div>
            </div>
            <Divider />
            <div className="delete-account-section">
                <p className="title">Delete your account</p>
                <label className="delete-account-description">
                    When you delete your account, you will lose access to Memphis,
                    <br />
                    and your profile will be permanently deleted. You can cancel the deletion for 14 days.
                </label>
                <div className="delete-account-checkbox">
                    <Checkbox
                        checked={checkboxdeleteAccount}
                        disabled={userName === 'root'}
                        onChange={() => setCheckboxdeleteAccount(!checkboxdeleteAccount)}
                        name="delete-account"
                    />
                    <p className={userName === 'root' ? 'disabled' : ''} onClick={() => userName !== 'root' && setCheckboxdeleteAccount(!checkboxdeleteAccount)}>
                        Confirm that I want to delete my account.
                    </p>
                </div>
                <Button
                    className="modal-btn"
                    width="200px"
                    height="36px"
                    placeholder="Delete Account"
                    colorType="white"
                    radiusType="circle"
                    backgroundColorType="red"
                    border="none"
                    boxShadowsType="red"
                    fontSize="14px"
                    fontWeight="600"
                    aria-haspopup="true"
                    disabled={!checkboxdeleteAccount}
                    onClick={() => modalFlip(true)}
                />
            </div>
            <Modal
                header="Remove accont"
                height="120px"
                rBtnText="Cancel"
                lBtnText="Remove"
                lBtnClick={() => {
                    removeMyUser();
                }}
                clickOutside={() => modalFlip(false)}
                rBtnClick={() => modalFlip(false)}
                open={open}
            >
                <label>Are you sure you want to remove user account?</label>
                <br />
                <label>Please note that this action is irreversible.</label>
            </Modal>
        </div>
    );
}

export default Profile;
