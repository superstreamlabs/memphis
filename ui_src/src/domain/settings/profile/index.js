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

import React, { useEffect, useContext, useState } from 'react';

import { LOCAL_STORAGE_ALLOW_ANALYTICS, LOCAL_STORAGE_USER_NAME } from '../../../const/localStorageConsts';
import { LOCAL_STORAGE_AVATAR_ID } from '../../../const/localStorageConsts';
import Button from '../../../components/button';
import { Context } from '../../../hooks/store';
import RadioButton from '../../../components/radioButton';
import { Checkbox } from 'antd';
import { Upload, message } from 'antd';
import ImgLoader from './imgLoader';

import pathDomains from '../../../router';
import { httpRequest } from '../../../services/http';
import { ApiEndpoints } from '../../../const/apiEndpoints';
import Modal from '../../../components/modal';
import { Divider } from 'antd';

function Profile() {
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

    const props = {
        name: 'file',
        action: 'https://www.mocky.io/v2/5cc8019d300000980a055e76',
        headers: {
            authorization: 'authorization-text'
        },

        onChange(info) {
            if (info.file.status !== 'uploading') {
                console.log(info.file, info.fileList);
            }

            if (info.file.status === 'done') {
                message.success(`${info.file.name} file uploaded successfully`);
            } else if (info.file.status === 'error') {
                message.error(`${info.file.name} file upload failed.`);
            }
        }
    };

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
            <div className="header">
                <p className="main-header">Profile</p>
                <p className="sub-header">Select your avater for showing in this system</p>
            </div>
            <div className="avatar-section">
                <p className="title">Avatar</p>
                <div className="avatar-images">
                    {Array.from(Array(9).keys()).map((item) => {
                        return (
                            <div
                                className={process.env.REACT_APP_SANDBOX_ENV ? 'sub-icon-wrapper-sandbox' : avatar === item + 1 ? 'avatar-img selected' : 'avatar-img'}
                                onClick={process.env.REACT_APP_SANDBOX_ENV ? '' : () => editAvatar(item + 1)}
                            >
                                <img
                                    src={localStorage.getItem('profile_pic') || require(`../../../assets/images/bots/avatar${item + 1}.svg`)} // profile_pic is available only in sandbox env
                                    width={localStorage.getItem('profile_pic') ? 35 : ''}
                                    height={localStorage.getItem('profile_pic') ? 35 : ''}
                                    alt="avater"
                                />
                            </div>
                        );
                    })}
                </div>
            </div>
            <div className="company-logo-section">
                <p className="title">Company Logo</p>
                <div className="company-logo">
                    <ImgLoader />
                    <div className="company-logo-right">
                        <div className="update-remove-logo">
                            <Button
                                className="modal-btn"
                                width="160px"
                                height="36px"
                                placeholder="Upload New"
                                colorType="white"
                                radiusType="circle"
                                backgroundColorType="purple"
                                fontSize="14px"
                                fontWeight="600"
                                aria-haspopup="true"
                                // onClick={() => modalFlip(true)}
                            />
                            <Button
                                className="modal-btn"
                                width="200px"
                                height="36px"
                                placeholder="Remove Current Logo"
                                colorType="red"
                                radiusType="circle"
                                backgroundColorType="none"
                                border="gray"
                                boxShadowsType="gray"
                                fontSize="14px"
                                fontWeight="600"
                                aria-haspopup="true"
                                // onClick={() => modalFlip(true)}
                            />
                        </div>
                        <label className="company-logo-description">Logo must be 200x200 pixel and size is less than 5mb</label>
                    </div>
                </div>
            </div>
            <Divider />
            <div className="analytics-section">
                <p className="title">Analytics</p>
                <label className="analytics-description">Lorem Ipsum is simply dummy text of the printing and typesetting industry.</label>
                <div className="radioButton-section">
                    <RadioButton
                        options={[
                            { id: 0, value: true, label: 'Allow Analytics' },
                            { id: 1, value: false, label: 'Don’t allow any analytics' }
                        ]}
                        radioValue={allowAnalytics}
                        onChange={(e) => sendAnalytics(e.target.value)}
                        labelType
                    />
                </div>
            </div>
            <Divider />
            <div className="delete-account-section">
                <p className="title">Delete your account</p>
                <label className="delete-account-description">
                    When you delete your account, you lose access to Front account services, and we permanently delete your personal data. You can cancel the deletion for
                    14 days.
                </label>
                <div className="delete-account-checkbox">
                    <Checkbox checked={checkboxdeleteAccount} onChange={() => setCheckboxdeleteAccount(!checkboxdeleteAccount)} name="delete-account" />
                    <p onClick={() => setCheckboxdeleteAccount(!checkboxdeleteAccount)}>Confirm that I want to delete my account.</p>
                </div>
                <Button
                    className="modal-btn"
                    width="200px"
                    height="36px"
                    placeholder="Delete Account"
                    colorType="white"
                    radiusType="circle"
                    backgroundColorType="red"
                    border="red"
                    boxShadowsType="red"
                    fontSize="14px"
                    fontWeight="600"
                    aria-haspopup="true"
                    disabled={!checkboxdeleteAccount || userName === 'root'}
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
