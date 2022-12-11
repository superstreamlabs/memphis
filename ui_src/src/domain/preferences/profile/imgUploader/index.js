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

import React, { useContext, useState } from 'react';
import { Context } from '../../../../hooks/store';
import { LOCAL_STORAGE_COMPANY_LOGO } from '../../../../const/localStorageConsts';
import { Upload, message } from 'antd';
import Button from '../../../../components/button';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import { httpRequest } from '../../../../services/http';
import Logo from '../../../../assets/images/logo.svg';

function ImgUploader() {
    const [state, dispatch] = useContext(Context);
    const [fileList, setFileList] = useState(
        localStorage.getItem(LOCAL_STORAGE_COMPANY_LOGO)
            ? [
                  {
                      uid: '1',
                      name: 'company_logo',
                      status: 'done',
                      url: localStorage.getItem(LOCAL_STORAGE_COMPANY_LOGO)
                  }
              ]
            : []
    );

    const props = {
        beforeUpload: (file) => {
            const isJpgOrPng = file.type === 'image/jpeg' || file.type === 'image/png';
            if (!isJpgOrPng) {
                message.error('JPG/PNG format required', 3);
            }
            setFileList([file]);
            return isJpgOrPng;
        },
        customRequest: (file) => uploadLogo(file),
        fileList
    };
    const uploadLogo = async ({ file, onSuccess, onError }) => {
        let dataImg = new FormData();
        dataImg.append('file', file);
        try {
            const data = await httpRequest('PUT', ApiEndpoints.EDIT_COMPANY_LOGO, dataImg);
            localStorage.setItem(LOCAL_STORAGE_COMPANY_LOGO, data.image);
            dispatch({ type: 'SET_COMPANY_LOGO', payload: data.image });
            onSuccess('ok');
        } catch (err) {
            onError('error');
        }
    };

    const deleteLogo = async ({ onSuccess, onError }) => {
        try {
            const data = await httpRequest('DELETE', ApiEndpoints.REMOVE_COMPANY_LOGO);
            localStorage.setItem(LOCAL_STORAGE_COMPANY_LOGO, null);
            dispatch({ type: 'SET_COMPANY_LOGO', payload: null });
            setFileList([]);
            onSuccess('ok');
        } catch (err) {
            onError('error');
        }
    };

    return (
        <div className="company-logo-section">
            <p className="title">Company Logo</p>
            <div className="company-logo">
                <img className="logoimg" src={state?.companyLogo || Logo} alt="companyLogo" />
                <div className="company-logo-right">
                    <div className="update-remove-logo">
                        <Upload {...props} name="company-logo" maxCount={1} showUploadList={false} fileList={fileList}>
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
                            />
                        </Upload>
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
                            onClick={() => deleteLogo(fileList[0])}
                            disabled={!state?.companyLogo}
                        />
                    </div>
                    <label className="company-logo-description">Logo must be 200x200 pixel and size is less than 5mb</label>
                </div>
            </div>
        </div>
    );
}

export default ImgUploader;
